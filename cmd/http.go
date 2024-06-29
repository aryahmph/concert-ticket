package cmd

import (
	"context"
	"github.com/aryahmph/concert-ticket/feature/category"
	"github.com/aryahmph/concert-ticket/feature/order"
	"github.com/aryahmph/concert-ticket/feature/shared"
	"github.com/aryahmph/concert-ticket/pkg"
	"github.com/midtrans/midtrans-go"
	"log"
	"net/http"
	"time"
)

func runHttpCmd(ctx context.Context) {
	cfg := shared.LoadConfig("config/http.yaml")

	// OpenTelemetry
	grpcConn := pkg.InitGrpcConn()
	shutdownTracerProvider := pkg.InitTracerProvider(ctx, grpcConn)

	// Dependency
	db := pkg.InitDatabase(ctx, cfg)
	defer db.Close()

	cache := pkg.InitCache(cfg)
	defer cache.Close()

	clientQueue := pkg.InitClientQueue(cfg)
	defer clientQueue.Close()

	inspectorQueue := pkg.InitInspectorQueue(cfg)
	defer inspectorQueue.Close()

	midtrans.ServerKey = cfg.Midtrans.ServerKey
	midtrans.Environment = midtrans.Sandbox

	category.SetDB(db)
	category.SetCache(cache)

	order.SetDB(db)
	order.SetCache(cache)
	order.SetQueueClient(clientQueue)
	order.SetQueueInspector(inspectorQueue)
	order.SetOrderCancellationDuration(time.Duration(cfg.Order.CancellationDuration) * time.Second)

	// Http server
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	category.HttpRoute(mux)
	order.HttpRoute(mux)

	srv := &http.Server{
		Addr:         cfg.Http.Addr(),
		Handler:      mux,
		ReadTimeout:  time.Duration(cfg.Http.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Http.WriteTimeout) * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalln("unable to start server", err)
		}
	}()

	log.Println("http server started")

	<-ctx.Done()

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalln("unable to shutdown server", err)
	}

	if err := shutdownTracerProvider(ctxShutDown); err != nil {
		log.Fatalf("failed to shutdown TracerProvider: %s", err)
	}

	log.Println("http server shutdown")
}
