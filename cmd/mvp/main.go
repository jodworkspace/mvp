package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/sessions"
	goredis "github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v2"
	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/internal/handler/rest"
	v1 "gitlab.com/jodworkspace/mvp/internal/handler/rest/v1"
	"gitlab.com/jodworkspace/mvp/internal/instrument"
	postgresrepo "gitlab.com/jodworkspace/mvp/internal/repository/postgres"
	"gitlab.com/jodworkspace/mvp/internal/usecase/oauth"
	taskuc "gitlab.com/jodworkspace/mvp/internal/usecase/task"
	useruc "gitlab.com/jodworkspace/mvp/internal/usecase/user"
	"gitlab.com/jodworkspace/mvp/pkg/db/postgres"
	"gitlab.com/jodworkspace/mvp/pkg/db/redis"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/monitor/metrics"
	"gitlab.com/jodworkspace/mvp/pkg/utils/cipherx"
	"google.golang.org/grpc"
)

func main() {
	initGob()
	cfg := config.LoadConfig()
	ctx := context.Background()

	metricsManager, err := newMetricsManager(cfg.Monitor.ServiceName, cfg.Monitor.CollectorEndpoint)
	if err != nil {
		panic(err)
	}

	shutdown, err := metricsManager.Init(ctx)
	defer shutdown(ctx)

	dbMetrics, err := metricsManager.NewPgxMetrics()
	if err != nil {
		panic(err)
	}

	httpMetrics, err := metricsManager.NewHTTPMetrics()
	if err != nil {
		panic(err)
	}

	zapLogger := logger.MustNewLogger(cfg.Logger.Level)
	aead := cipherx.MustNewAEAD([]byte(cfg.Server.AESKey))

	pgClient, err := postgres.NewPostgresConnection(
		cfg.Postgres.DSN(),
		postgres.WithMaxConns(10),
		postgres.WithMinConns(2),
	)
	if err != nil {
		panic(err)
	}
	defer pgClient.Pool().Close()

	pgInstrumentedClient := instrument.NewInstrumentedClient(pgClient, dbMetrics)

	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	rdb, err := redis.NewClient(redisClient)
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	sessionStore := redis.NewStore(rdb, "session:", &sessions.Options{
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

			transactionManager := postgresrepo.NewTransactionManager(pgClient)

			taskRepository := postgresrepo.NewTaskRepository(pgInstrumentedClient)
			taskUC := taskuc.NewTaskUseCase(taskRepository, zapLogger)
			taskHandler := v1.NewTaskHandler(taskUC, zapLogger)

			userRepository := postgresrepo.NewUserRepository(pgClient)
			linkRepository := postgresrepo.NewLinkRepository(pgClient)
			userUC := useruc.NewUserUseCase(userRepository, linkRepository, transactionManager, aead, zapLogger)

			googleUC := oauthuc.NewGoogleUseCase(cfg.GoogleOAuth, zapLogger)
			oauthMng := oauthuc.NewManager(cfg.Token, zapLogger)
			oauthMng.RegisterOAuthProvider(googleUC)

			oauthHandler := v1.NewOAuthHandler(sessionStore, userUC, oauthMng, zapLogger)

			srv := rest.NewServer(cfg, sessionStore, taskHandler, oauthHandler, zapLogger, httpMetrics)
			srv.Run()

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func initGob() {
	gob.Register(&domain.User{})
	gob.Register(&domain.Document{})
}

func newMetricsManager(serviceName, endpoint string) (*metrics.Manager, error) {
	conn, err := grpc.NewClient(endpoint)
	if err != nil {
		return nil, err
	}

	return metrics.NewManager(serviceName, conn), nil
}
