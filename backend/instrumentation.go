package main

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

func NewOtelProviders(ctx context.Context) (*trace.TracerProvider, *log.LoggerProvider, error) {
	tracerProvider, err := newTracerProvider(ctx)
	if err != nil {
		return nil, nil, err
	}

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	)
	loggerProvider, err := newLoggerProvider(ctx)
	if err != nil {
		return nil, nil, err
	}
	global.SetLoggerProvider(loggerProvider)
	return tracerProvider, loggerProvider, nil
}

func newTracerProvider(ctx context.Context) (*trace.TracerProvider, error) {
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, err
	}
	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, err
	}

	bsp := trace.NewBatchSpanProcessor(traceExporter)

	provider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
		)),
	)

	return provider, nil
}

func newLoggerProvider(ctx context.Context) (*log.LoggerProvider, error) {
	exporter, err := otlploggrpc.New(ctx)
	if err != nil {
		return nil, err
	}
	processor := log.NewBatchProcessor(exporter)
	provider := log.NewLoggerProvider(
		log.WithProcessor(processor),
		log.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
		)),
	)
	return provider, nil
}
