package main

import (
	"log"

	"nats-streaming/config"
	"nats-streaming/internal/server"
	"nats-streaming/pkg/jaeger"
	"nats-streaming/pkg/logger"
	"nats-streaming/pkg/nats"
	"nats-streaming/pkg/postgresql"
	"nats-streaming/pkg/redis"

	"github.com/opentracing/opentracing-go"
)

// @title Email microservice
// @version 1.0
// @description Email microservice
// @termsOfService http://swagger.io/terms/

// @contact.name huylq
// @contact.url https://github.com/huylq
// @contact.email lequochuy9302gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:5000
// @BasePath /api/v1
func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.NewApiLogger(cfg)
	appLogger.InitLogger()
	appLogger.Info("Starting emails microservice")
	appLogger.Infof(
		"AppVersion: %s, LogLevel: %s, DevelopmentMode: %s",
		cfg.AppVersion,
		cfg.Logger.Level,
		cfg.HTTP.Development,
	)
	appLogger.Infof("Success loaded config: %+v", cfg.AppVersion)

	tracer, closer, err := jaeger.InitJaeger(cfg)
	if err != nil {
		appLogger.Fatal("cannot create tracer", err)
	}
	appLogger.Info("Jaeger connected")

	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	appLogger.Info("Opentracing connected")

	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		appLogger.Fatalf("NewRedisClient: %+v", err)
	}

	appLogger.Infof("Redis connected: %+v", redisClient.PoolStats())

	natsConn, err := nats.NewNatsConnect(cfg, appLogger)
	if err != nil {
		appLogger.Fatalf("NewNatsConnect: %+v", err)
	}
	appLogger.Infof(
		"Nats Connected: Status: %+v IsConnected: %v ConnectedUrl: %v ConnectedServerId: %v",
		natsConn.NatsConn().Status(),
		natsConn.NatsConn().IsConnected(),
		natsConn.NatsConn().ConnectedUrl(),
		natsConn.NatsConn().ConnectedServerId(),
	)

	pgxPool, err := postgresql.NewPgxConn(cfg)
	if err != nil {
		appLogger.Fatalf("NewPgxConn: %+v", err)
	}
	appLogger.Infof("PostgreSQL connected: %+v", pgxPool.Stat().TotalConns())

	s := server.NewServer(appLogger, cfg, natsConn, pgxPool, tracer, redisClient)

	appLogger.Fatal(s.Run())
}
