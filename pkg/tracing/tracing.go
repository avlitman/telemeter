package tracing

import (
	"context"
	"fmt"

	propjaeger "go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

// Adjusted from github.com/observatorium/api/tracing/tracing.go

// EndpointType represents the type of the tracing endpoint.
type EndpointType string

const (
	EndpointTypeCollector EndpointType = "collector"
	EndpointTypeAgent     EndpointType = "agent"
)

// InitTracer creates an OTel TracerProvider that exports the traces to a Jaeger agent/collector.
func InitTracer(
	serviceName string,
	endpoint string,
	endpointTypeRaw string,
	samplingFraction float64,
) (tp trace.TracerProvider, err error) {
	if endpoint == "" {
		return trace.NewNoopTracerProvider(), nil
	}

	endpointOption := jaeger.WithAgentEndpoint(
		jaeger.WithAgentHost(endpoint),
	)
	if EndpointType(endpointTypeRaw) == EndpointTypeCollector {
		endpointOption = jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(endpoint),
		)
	}

	exp, err := jaeger.NewRawExporter(
		endpointOption,
	)
	if err != nil {
		return tp, fmt.Errorf("create jaeger export pipeline: %w", err)
	}

	r, err := resource.New(context.Background(), resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)))
	if err != nil {
		return tp, fmt.Errorf("create resource: %w", err)
	}

	tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(samplingFraction)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propjaeger.Jaeger{},
		propagation.Baggage{},
	))

	return tp, nil
}
