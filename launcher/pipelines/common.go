package pipelines

import (
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// Protocol defines the possible values of the protocol field.
type Protocol string

// These are the only possible values for Protocol.
const (
	ProtocolGRPC         Protocol = "grpc"
	ProtocolHTTPProtobuf Protocol = "http/protobuf"
	ProtocolHTTPJSON     Protocol = "http/json"
)

// PipelineConfig contains config info for a Pipeline.
type PipelineConfig struct {
	Protocol        Protocol
	Endpoint        string
	Insecure        bool
	Headers         map[string]string
	Resource        *resource.Resource
	ReportingPeriod string
	Propagators     []string
	SpanProcessors  []trace.SpanProcessor
	Sampler         trace.Sampler
}

// PipelineSetupFunc defines the interface for a Pipeline Setup function.
type PipelineSetupFunc func(PipelineConfig) (func() error, error)
