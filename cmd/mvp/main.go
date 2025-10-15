package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/urfave/cli/v2"
	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/internal/handler/rest"
	v1 "gitlab.com/jodworkspace/mvp/internal/handler/rest/v1"
	pgrepo "gitlab.com/jodworkspace/mvp/internal/repository/postgres"
	"gitlab.com/jodworkspace/mvp/internal/usecase/document"
	"gitlab.com/jodworkspace/mvp/internal/usecase/oauth"
	"gitlab.com/jodworkspace/mvp/internal/usecase/task"
	"gitlab.com/jodworkspace/mvp/internal/usecase/user"
	"gitlab.com/jodworkspace/mvp/pkg/db/postgres"
	"gitlab.com/jodworkspace/mvp/pkg/db/redis"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/otel"
	otelhttp "gitlab.com/jodworkspace/mvp/pkg/otel/http"
	otelpgx "gitlab.com/jodworkspace/mvp/pkg/otel/pgx"
	"gitlab.com/jodworkspace/mvp/pkg/utils/cipherx"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	initGob()
	cfg := config.LoadConfig()
	ctx := context.Background()

	conn, err := grpc.NewClient(cfg.Monitor.CollectorEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	panicOnErr(err)
	defer conn.Close()

	otelManager := otel.NewManager(&otel.Config{
		ServiceName:    cfg.Monitor.ServiceName,
		MetricInterval: cfg.Monitor.MetricInterval,
	}, otel.WithGRPCConn(conn), otel.WithCustomPrometheus())

	shutdown, err := otelManager.SetupOtelSDK(ctx)
	panicOnErr(err)
	defer shutdown(ctx)

	pgxMonitor, err := otelpgx.NewMonitor()
	panicOnErr(err)

	zapLogger := logger.MustNewLogger(cfg.Logger.Level)
	aead := cipherx.MustNewAEAD([]byte(cfg.Server.AESKey))

	pgClient, err := postgres.NewPostgresDB(
		cfg.Postgres.DSN(),
		postgres.WithMaxConns(10),
		postgres.WithMinConns(2),
		postgres.WithQueryTrace(pgxMonitor),
	)
	panicOnErr(err)
	defer pgClient.Pool().Close()

	redisClient, err := redis.NewClient(ctx,
		fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		redis.WithCredential(cfg.Redis.Username, cfg.Redis.Password),
		redis.WithDB(cfg.Redis.DB),
	)
	panicOnErr(err)
	defer redisClient.Close()

	err = redisClient.Instrument()
	panicOnErr(err)

	sessionStore := redis.NewStore(redisClient, "session:", &sessions.Options{
		Path:     cfg.Session.CookiePath,
		Domain:   cfg.Session.Domain,
		MaxAge:   cfg.Session.MaxAge,
		HttpOnly: cfg.Session.HTTPOnly,
		Secure:   cfg.Session.Secure,
	})

	httpClient := httpx.NewHTTPClient(http.Client{
		Timeout:   10 * time.Second,
		Transport: otelhttp.TransportWithTracing(),
	})

	app := &cli.App{
		Name:  "Jod",
		Usage: "MVP Server",
		Action: func(c *cli.Context) error {
			// DB Transaction
			transactionManager := pgrepo.NewTransactionManager(pgClient)

			// Tasks
			taskRepository := pgrepo.NewTaskRepository(pgClient)
			taskUC := task.NewUseCase(taskRepository, zapLogger)
			taskHandler := v1.NewTaskHandler(taskUC, zapLogger)

			// Users
			userRepository := pgrepo.NewUserRepository(pgClient)
			linkRepository := pgrepo.NewLinkRepository(pgClient)
			userUC := user.NewUseCase(userRepository, linkRepository, transactionManager, aead, zapLogger)

			// OAuth
			googleUC := oauth.NewGoogleUseCase(cfg.GoogleOAuth, httpClient, zapLogger)
			oauthMng := oauth.NewManager(cfg.Token, zapLogger)
			oauthMng.RegisterOAuthProvider(googleUC)
			oauthHandler := v1.NewOAuthHandler(sessionStore, userUC, oauthMng, zapLogger)

			documentUC := document.NewUseCase(httpClient, zapLogger)
			documentHandler := v1.NewDocumentHandler(documentUC, zapLogger)
			wsHandler := v1.NewWSHandler(documentUC, zapLogger)

			// Start server
			srv := rest.NewServer(
				cfg,
				aead,
				sessionStore,
				taskHandler,
				oauthHandler,
				documentHandler,
				wsHandler,
				zapLogger,
				otelManager,
			)

			return srv.Run()
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func initGob() {
	gob.Register(&domain.User{})
	gob.Register(&domain.Document{})
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
