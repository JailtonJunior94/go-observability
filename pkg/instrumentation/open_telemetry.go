package instrumentation

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type Instrumentation struct {
	tracer         trace.Tracer
	meterProvider  *metric.MeterProvider
	tracerProvider *sdktrace.TracerProvider
}

func NewInstrumentation(ctx context.Context, serviceName, serviceVersion string, endpoint string) *Instrumentation {
	resource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	mp := newMeterProviderPrometheus(ctx, resource, endpoint)
	otel.SetMeterProvider(mp)

	tp := newTraceProvider(ctx, resource, endpoint)
	otel.SetTracerProvider(tp)

	return &Instrumentation{
		meterProvider:  mp,
		tracerProvider: tp,
		tracer:         tp.Tracer(serviceName),
	}
}

func newTraceProvider(ctx context.Context, resource *resource.Resource, endpoint string) *sdktrace.TracerProvider {
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
	)

	if err != nil {
		log.Fatal(err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(resource),
	)
	return tracerProvider
}

func newMeterProvider(ctx context.Context, resource *resource.Resource, endpoint string) *metric.MeterProvider {
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		log.Fatal(err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(resource),
		metric.WithReader(metric.NewPeriodicReader(
			metricExporter,
			metric.WithInterval(3*time.Second)),
		),
	)
	return meterProvider
}

func newMeterProviderPrometheus(ctx context.Context, resource *resource.Resource, endpoint string) *metric.MeterProvider {
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}
	meterProvider := metric.NewMeterProvider(metric.WithReader(exporter))
	return meterProvider
}

func (i *Instrumentation) GetTracer() trace.Tracer {
	return i.tracer
}

func (i *Instrumentation) MeterProvider() *metric.MeterProvider {
	return i.meterProvider
}

func (i *Instrumentation) TracerProvider() *sdktrace.TracerProvider {
	return i.tracerProvider
}
