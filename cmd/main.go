package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/jailtonjunior94/go-instrumentation/pkg/http/middlewares"
	"github.com/jailtonjunior94/go-instrumentation/pkg/observability"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx := context.Background()

	observability := observability.NewObservability(
		observability.WithServiceName("go-telemetry"),
		observability.WithServiceVersion("1.0.0"),
		observability.WithResource(),
		observability.WithTracerProvider(ctx, "localhost:4317"),
		observability.WithMeterProvider(ctx, "localhost:4317"),
	)

	tracerProvider := observability.TracerProvider()
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	meterProvider := observability.MeterProvider()
	defer func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	metricsMiddleware, err := middlewares.NewHTTPMetricsMiddleware(observability)
	if err != nil {
		log.Fatal(err)
	}

	router := chi.NewRouter()
	router.Use(
		middleware.Logger,
		middleware.Recoverer,
		metricsMiddleware.Metrics,
		middleware.Heartbeat("/health"),
		middleware.SetHeader("Content-Type", "application/json"),
	)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	server := http.Server{
		ReadTimeout:       time.Duration(10) * time.Second,
		ReadHeaderTimeout: time.Duration(10) * time.Second,
		Handler:           router,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", "7001"))
	if err != nil {
		log.Fatal(err)
	}
	server.Serve(listener)
}
