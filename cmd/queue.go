package cmd

import (
	"context"
	"github.com/aryahmph/concert-ticket/feature/order"
	"github.com/aryahmph/concert-ticket/feature/shared"
	"github.com/aryahmph/concert-ticket/pkg"
	"github.com/hibiken/asynq"
	"log"
	"time"
)

func runQueueCmd(ctx context.Context) {
	cfg := shared.LoadConfig("config/queue.yaml")

	// OpenTelemetry
	grpcConn := pkg.InitGrpcConn()
	shutdownTracerProvider := pkg.InitTracerProvider(ctx, grpcConn)

	// Dependency
	db := pkg.InitDatabase(ctx, cfg)
	defer db.Close()

	cache := pkg.InitCache(cfg)
	defer cache.Close()

	order.SetDB(db)
	order.SetCache(cache)

	server := asynq.NewServer(asynq.RedisClientOpt{Addr: cfg.Cache.Addr, PoolSize: int(cfg.Cache.MinConn)},
		asynq.Config{
			LogLevel:        asynq.WarnLevel,
			Concurrency:     int(cfg.Queue.Concurrent),
			ShutdownTimeout: 5 * time.Second,
		})

	mux := asynq.NewServeMux()
	mux.HandleFunc(order.CancellationTaskName, order.CancellationHandler)

	go func() {
		err := server.Run(mux)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	log.Println("queue started")

	<-ctx.Done()

	server.Shutdown()

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := shutdownTracerProvider(ctxShutDown); err != nil {
		log.Fatalf("failed to shutdown TracerProvider: %s", err)
	}

	log.Println("queue stopped")
}
