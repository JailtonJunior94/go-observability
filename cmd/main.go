package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jailtonjunior94/go-instrumentation/pkg/instrumentation"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func main() {
	ctx := context.Background()
	instrumentation := instrumentation.NewInstrumentation(ctx, "sample", "1.0.0", "localhost:4317")

	tracer := instrumentation.GetTracer()
	tp := instrumentation.TracerProvider()
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	mp := instrumentation.MeterProvider()
	defer func() {
		if err := mp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	router := chi.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	router.Get("/api", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "roll")
		defer span.End()

		meter := mp.Meter("sample")
		rollCnt, _ := meter.Int64Counter("dice.rolls",
			metric.WithDescription("The number of rolls by roll value"),
			metric.WithUnit("{roll}"),
		)

		roll := 1 + rand.Intn(6)
		rollValueAttr := attribute.Int("roll.value", roll)
		rollCnt.Add(ctx, 1, metric.WithAttributes(rollValueAttr))

		w.WriteHeader(http.StatusOK)
		fmt.Println(ctx)
	})

	server := http.Server{
		ReadTimeout:       time.Duration(10) * time.Second,
		ReadHeaderTimeout: time.Duration(10) * time.Second,
		Handler:           router,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", "9000"))
	if err != nil {
		panic(err)
	}
	server.Serve(listener)
}
