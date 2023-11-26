package pipelines

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"

	hostMetrics "go.opentelemetry.io/contrib/instrumentation/host"
	runtimeMetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
)

// NewMetricsPipeline takes a PipelineConfig and builds a metrics pipeline.
// It returns a shutdown function that should be called when terminating the pipeline.
func NewMetricsPipeline(c PipelineConfig) (func() error, error) {
	metricExporter, err := newMetricsExporter(c.Protocol, c.Endpoint, c.Insecure, c.Headers)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %v", err)
	}

	var readerOpts []metric.PeriodicReaderOption
	if c.ReportingPeriod != "" {
		period, err := time.ParseDuration(c.ReportingPeriod)
		if err != nil {
			return nil, fmt.Errorf("invalid metric reporting period: %v", err)
		}
		if period <= 0 {
			return nil, fmt.Errorf("invalid metric reporting period: %v", c.ReportingPeriod)
		}
		readerOpts = append(readerOpts, metric.WithInterval(period))
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(c.Resource),
		metric.WithReader(metric.NewPeriodicReader(metricExporter, readerOpts...)))

	if err = runtimeMetrics.Start(runtimeMetrics.WithMeterProvider(meterProvider)); err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics: %v", err)
	}

	if err = hostMetrics.Start(hostMetrics.WithMeterProvider(meterProvider)); err != nil {
		return nil, fmt.Errorf("failed to start host metrics: %v", err)
	}

	otel.SetMeterProvider(meterProvider)
	return func() error {
		return meterProvider.Shutdown(context.Background())
	}, nil
}

//revive:disable:flag-parameter bools are fine for an internal function
func newMetricsExporter(protocol Protocol, endpoint string, insecure bool, headers map[string]string) (metric.Exporter, error) {
	switch protocol {
	case "grpc":
		return newGRPCMetricsExporter(endpoint, insecure, headers)
	case "http/protobuf":
		return newHTTPMetricsExporter(endpoint, insecure, headers)
	case "http/json":
		return nil, errors.New("http/json is currently unsupported")
	default:
		return nil, errors.New("'" + string(protocol) + "' is not a supported protocol")
	}
}

func newGRPCMetricsExporter(endpoint string, insecure bool, headers map[string]string) (metric.Exporter, error) {
	secureOption := otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if insecure {
		secureOption = otlpmetricgrpc.WithInsecure()
	}
	return otlpmetricgrpc.New(
		context.Background(),
		secureOption,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithHeaders(headers),
		otlpmetricgrpc.WithCompressor(gzip.Name),
	)
}

func newHTTPMetricsExporter(endpoint string, insecure bool, headers map[string]string) (metric.Exporter, error) {
	tlsconfig := &tls.Config{}
	secureOption := otlpmetrichttp.WithTLSClientConfig(tlsconfig)
	if insecure {
		secureOption = otlpmetrichttp.WithInsecure()
	}
	return otlpmetrichttp.New(
		context.Background(),
		secureOption,
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithHeaders(headers),
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
	)
}
