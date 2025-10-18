package pipelines

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

func NewLogsPipeline(c PipelineConfig) (func() error, error) {
	opts := []log.LoggerProviderOption{
		log.WithResource(c.Resource),
	}

	logsExporter, err := newLogsExporter(c.Protocol, c.Endpoint, c.Insecure, c.Headers)
	if err != nil {
		return nil, fmt.Errorf("failed to create logs exporter: %v", err)
	}

	blp := log.NewBatchProcessor(logsExporter)
	opts = append(opts, log.WithProcessor(blp))

	lp := log.NewLoggerProvider(opts...)

	global.SetLoggerProvider(lp)

	return func() error {
		_ = blp.Shutdown(context.Background())
		return logsExporter.Shutdown(context.Background())
	}, nil
}

func newLogsExporter(protocol Protocol, endpoint string, insecure bool, headers map[string]string) (log.Exporter, error) {
	switch protocol {
	case ProtocolGRPC:
		return newGRPCLoggerExporter(endpoint, insecure, headers)
	case ProtocolHTTPProtobuf:
		return newHTTPLoggerExporter(endpoint, insecure, headers)
	case ProtocolHTTPJSON:
		return nil, errors.New("http/json is currently unsupported")
	default:
		return nil, errors.New("'" + string(protocol) + "' is not a supported protocol")
	}
}

func newGRPCLoggerExporter(endpoint string, insecure bool, headers map[string]string) (log.Exporter, error) {
	secureOption := otlploggrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if insecure {
		secureOption = otlploggrpc.WithInsecure()
	}
	return otlploggrpc.New(
		context.Background(),
		secureOption,
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithHeaders(headers),
		otlploggrpc.WithCompressor(gzip.Name),
	)
}

func newHTTPLoggerExporter(endpoint string, insecure bool, headers map[string]string) (log.Exporter, error) {
	tlsConfig := &tls.Config{}
	secureOption := otlploghttp.WithTLSClientConfig(tlsConfig)
	if insecure {
		secureOption = otlploghttp.WithInsecure()
	}
	return otlploghttp.New(
		context.Background(),
		secureOption,
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithHeaders(headers),
		otlploghttp.WithCompression(otlploghttp.GzipCompression),
	)
}
