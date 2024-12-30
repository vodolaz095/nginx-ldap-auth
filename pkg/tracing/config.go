package tracing

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type Config struct {
	Endpoint string  `yaml:"endpoint" validate:"hostname_port"`
	Ratio    float64 `yaml:"ratio" validate:"lte=1,gte=0"`
}

const defaultTimeout = 10 * time.Second

var exp *jaeger.Exporter

func Configure(cfg Config, extraAttributes ...attribute.KeyValue) (err error) {
	if cfg.Ratio == 0 {
		return nil
	}
	parts := strings.Split(cfg.Endpoint, ":")
	if len(parts) != 2 {
		return fmt.Errorf("malformed endpoint: %s", cfg.Endpoint)
	}

	// export via compact thrift protocol over upd - important
	exp, err = jaeger.New(jaeger.WithAgentEndpoint(
		jaeger.WithAgentHost(parts[0]),
		jaeger.WithAgentPort(parts[1]),
	))
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	extraAttributes = append(extraAttributes,
		semconv.HostID(hostname),
	)
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// set sampling part of data
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(cfg.Ratio)),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			extraAttributes...,
		)),
	)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	return nil
}

func Wait(ctx context.Context) (err error) {
	<-ctx.Done()

	if exp != nil {
		shutdownContext, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		return exp.Shutdown(shutdownContext)
	}

	return nil
}
