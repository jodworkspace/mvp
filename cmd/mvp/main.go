package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	goredis "github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v2"
	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/internal/handler/rest"
	v1 "gitlab.com/jodworkspace/mvp/internal/handler/rest/v1"
	postgresrepo "gitlab.com/jodworkspace/mvp/internal/repository/postgres"
	"gitlab.com/jodworkspace/mvp/internal/usecase/oauth"
	taskuc "gitlab.com/jodworkspace/mvp/internal/usecase/task"
	useruc "gitlab.com/jodworkspace/mvp/internal/usecase/user"
	"gitlab.com/jodworkspace/mvp/pkg/db/postgres"
	"gitlab.com/jodworkspace/mvp/pkg/db/redis"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/otel"
	otelhttp "gitlab.com/jodworkspace/mvp/pkg/otel/http"
	otelpgx "gitlab.com/jodworkspace/mvp/pkg/otel/pgx"
	"gitlab.com/jodworkspace/mvp/pkg/utils/cipherx"
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
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	otelManager := otel.NewManager(&otel.Config{
		ServiceName:    cfg.Monitor.ServiceName,
		MetricInterval: cfg.Monitor.MetricInterval,
	}, otel.WithGRPCConn(conn), otel.WithCustomPrometheus())

	shutdown, err := otelManager.SetupOtelSDK(ctx)
	if err != nil {
		panic(err)
	}
	defer shutdown(ctx)

	pgxMonitor, err := otelpgx.NewMonitor()
	if err != nil {
		panic(err)
	}

	httpMonitor, err := otelhttp.NewMonitor()
	if err != nil {
		panic(err)
	}

	zapLogger := logger.MustNewLogger(cfg.Logger.Level)
	aead := cipherx.MustNewAEAD([]byte(cfg.Server.AESKey))

	pgClient, err := postgres.NewPostgresConnection(
		cfg.Postgres.DSN(),
		postgres.WithMaxConns(10),
		postgres.WithMinConns(2),
		postgres.WithQueryTrace(pgxMonitor), // TODO: Add metrics + tracer
	)
	if err != nil {
		panic(err)
	}
	defer pgClient.Pool().Close()

	redisClient, err := redis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	sessionStore := redis.NewStore(redisClient, "session:", &sessions.Options{
		Path:     cfg.Session.CookiePath,
		Domain:   cfg.Session.Domain,
		MaxAge:   cfg.Session.MaxAge,
		HttpOnly: cfg.Session.HTTPOnly,
		Secure:   cfg.Session.Secure,
	})

	app := &cli.App{
		Name:  "Jod",
		Usage: "MVP Server",
		Action: func(c *cli.Context) error {
			// Instrumented clients
			httpClient := http.Client{
				Transport: httpMonitor.TransportWithTracing(),
			}

			// DB Transaction
			transactionManager := postgresrepo.NewTransactionManager(pgClient)

			// Tasks
			taskRepository := postgresrepo.NewTaskRepository(pgClient)
			taskUC := taskuc.NewTaskUseCase(taskRepository, zapLogger)
			taskHandler := v1.NewTaskHandler(taskUC, zapLogger)

			// Users
			userRepository := postgresrepo.NewUserRepository(pgClient)
			linkRepository := postgresrepo.NewLinkRepository(pgClient)
			userUC := useruc.NewUserUseCase(userRepository, linkRepository, transactionManager, aead, zapLogger)

			// OAuth
			googleUC := oauthuc.NewGoogleUseCase(cfg.GoogleOAuth, httpClient, zapLogger)
			oauthMng := oauthuc.NewManager(cfg.Token, zapLogger)
			oauthMng.RegisterOAuthProvider(googleUC)
			oauthHandler := v1.NewOAuthHandler(sessionStore, userUC, oauthMng, zapLogger)

			// Start server
			srv := rest.NewServer(cfg, sessionStore, taskHandler, oauthHandler, zapLogger, otelManager, httpMonitor)
			srv.Run()

			return nil
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
