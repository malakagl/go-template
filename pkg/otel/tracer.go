package otel

import (
	"context"
	"log"

	"github.com/malakagl/kart-challenge/pkg/constants"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var service string

func SetServiceName(s string) {
	service = s
}

func InitTracer(serviceName, otlpEndpoint string) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()
	SetServiceName(serviceName)
	// Create OTLP HTTP exporter
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(otlpEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Create resource with service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	// Set global TracerProvider
	otel.SetTracerProvider(tp)

	log.Printf("Tracer initialized for service: %s -> sending traces to OTLP endpoint %s", serviceName, otlpEndpoint)
	return tp, nil
}

func Tracer(ctx context.Context, s string) (context.Context, trace.Span) {
	sCtx, span := otel.Tracer(service).Start(ctx, s)
	parentSpan := trace.SpanFromContext(ctx)
	if parentSpan != nil {
		sCtx = context.WithValue(sCtx, constants.ParentSpanId, parentSpan.SpanContext().SpanID().String())
	}
	return sCtx, span
}
