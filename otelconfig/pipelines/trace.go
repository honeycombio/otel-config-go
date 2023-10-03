package pipelines

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

// NewTracePipeline creates a new trace pipeline from a config.
// It returns a shutdown function that should be called when terminating the pipeline.
func NewTracePipeline(c PipelineConfig) (func() error, error) {
	opts := []trace.TracerProviderOption{
		trace.WithResource(c.Resource),
		trace.WithSampler(c.Sampler),
	}
	for _, sp := range c.SpanProcessors {
		opts = append(opts, trace.WithSpanProcessor(sp))
	}

	shutdown := emptyShutdown
	if !c.DisableDefaultSpanProcessor {
		// make sure the exporter is added last
		spanExporter, err := newTraceExporter(c.Protocol, c.Endpoint, c.Insecure, c.Headers)
		if err != nil {
			return nil, fmt.Errorf("failed to create span exporter: %v", err)
		}

		bsp := trace.NewBatchSpanProcessor(spanExporter)
		opts = append(opts, trace.WithSpanProcessor(bsp))

		shutdown = func() error {
			_ = bsp.Shutdown(context.Background())
			return spanExporter.Shutdown(context.Background())
		}
	} else if len(c.SpanProcessors) == 0 {
		return nil, fmt.Errorf("must provide at least one span processor if the default span processor is disabled")
	}

	tp := trace.NewTracerProvider(opts...)
	if err := configurePropagators(c); err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tp)

	return shutdown, nil
}

//revive:disable:flag-parameter bools are fine for an internal function
func newTraceExporter(protocol Protocol, endpoint string, insecure bool, headers map[string]string) (*otlptrace.Exporter, error) {
	switch protocol {
	case "grpc":
		return newGRPCTraceExporter(endpoint, insecure, headers)
	case "http/protobuf":
		return newHTTPTraceExporter(endpoint, insecure, headers)
	case "http/json":
		return nil, errors.New("http/json is currently unsupported")
	default:
		return nil, errors.New("'" + string(protocol) + "' is not a supported protocol")
	}
}

func newGRPCTraceExporter(endpoint string, insecure bool, headers map[string]string) (*otlptrace.Exporter, error) {
	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if insecure {
		secureOption = otlptracegrpc.WithInsecure()
	}
	return otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithHeaders(headers),
			otlptracegrpc.WithCompressor(gzip.Name),
		),
	)
}

func newHTTPTraceExporter(endpoint string, insecure bool, headers map[string]string) (*otlptrace.Exporter, error) {
	tlsconfig := &tls.Config{}
	secureOption := otlptracehttp.WithTLSClientConfig(tlsconfig)
	if insecure {
		secureOption = otlptracehttp.WithInsecure()
	}
	return otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(
			secureOption,
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithHeaders(headers),
			otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
		),
	)
}

// configurePropagators configures B3 propagation by default.
func configurePropagators(c PipelineConfig) error {
	propagatorsMap := map[string]propagation.TextMapPropagator{
		"b3":           b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)),
		"baggage":      propagation.Baggage{},
		"tracecontext": propagation.TraceContext{},
		"ottrace":      ot.OT{},
	}
	var props []propagation.TextMapPropagator
	for _, key := range c.Propagators {
		prop := propagatorsMap[key]
		if prop != nil {
			props = append(props, prop)
		}
	}
	if len(props) == 0 {
		return fmt.Errorf("invalid configuration: unsupported propagators. Supported options: b3,baggage,tracecontext,ottrace")
	}
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		props...,
	))
	return nil
}

func emptyShutdown() error {
	return nil
}
