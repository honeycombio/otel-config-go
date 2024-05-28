package otelconfig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"go.opentelemetry.io/contrib/detectors/aws/lambda"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	collectormetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

//revive:disable:import-shadowing this is a test file

const (
	expectedTracingDisabledMessage       = "tracing is disabled by configuration: no endpoint set"
	expectedMetricsDisabledMessage       = "metrics are disabled by configuration: no endpoint set"
	expectedTracingDisabledConfigMessage = "tracing is disabled by configuration: enabled set to false"
	expectedMetricsDisabledConfigMessage = "metrics are disabled by configuration: enabled set to false"
)

type testLogger struct {
	output []string
}

func (logger *testLogger) addOutput(output string) {
	logger.output = append(logger.output, output)
}

func (logger *testLogger) Fatalf(format string, v ...interface{}) {
	logger.addOutput(fmt.Sprintf(format, v...))
}

func (logger *testLogger) Debugf(format string, v ...interface{}) {
	logger.addOutput(fmt.Sprintf(format, v...))
}

func (logger *testLogger) requireContains(t *testing.T, expected string) {
	t.Helper()
	for _, output := range logger.output {
		if strings.Contains(output, expected) {
			return
		}
	}

	t.Errorf("\nString unexpectedly not found: %v\nIn: %v", expected, logger.output)
}

func (logger *testLogger) requireNotContains(t *testing.T, expected string) {
	t.Helper()
	for _, output := range logger.output {
		if strings.Contains(output, expected) {
			t.Errorf("\nString unexpectedly found: %v\nIn: %v", expected, logger.output)
			return
		}
	}
}

var trueVal = true
var falseVal = false

// Create some dummy server implementations so that we can
// spin up tests that don't need to wait for a timeout trying to send data.
type dummyTraceServer struct {
	collectortrace.UnimplementedTraceServiceServer

	recievedExportTraceServiceRequests []*collectortrace.ExportTraceServiceRequest
}

func (s *dummyTraceServer) Export(ctx context.Context, req *collectortrace.ExportTraceServiceRequest) (*collectortrace.ExportTraceServiceResponse, error) {
	s.recievedExportTraceServiceRequests = append(s.recievedExportTraceServiceRequests, req)

	return &collectortrace.ExportTraceServiceResponse{}, nil
}

type dummyMetricsServer struct {
	collectormetrics.UnimplementedMetricsServiceServer
}

func (*dummyMetricsServer) Export(ctx context.Context, req *collectormetrics.ExportMetricsServiceRequest) (*collectormetrics.ExportMetricsServiceResponse, error) {
	return &collectormetrics.ExportMetricsServiceResponse{}, nil
}

// dummyGRPCListener is a test helper that builds a dummy grpc server that does nothing but
// returns quickly so that we don't have to wait for timeouts.
func dummyGRPCListener() func() {
	return dummyGRPCListenerWithTraceServer(&dummyTraceServer{})
}

func dummyGRPCListenerWithTraceServer(traceServer collectortrace.TraceServiceServer) func() {
	grpcServer := grpc.NewServer()
	collectortrace.RegisterTraceServiceServer(grpcServer, traceServer)
	collectormetrics.RegisterMetricsServiceServer(grpcServer, &dummyMetricsServer{})

	// we listen on localhost, not 0.0.0.0, because otherwise firewalls can get upset
	// and get in the way of testing.
	l, err := net.Listen("tcp", net.JoinHostPort("localhost", "4317"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		panic("oops - dummyGrpcListener failed to start up!")
	}
	go func() {
		_ = grpcServer.Serve(l)
	}()
	return grpcServer.Stop
}

// withTestExporters conforms to the Option interface and sets up the options needed
// to prevent a test from having to time out. It won't work unless the test also does this:
//
// stopper := dummyGRPCListener()
// defer stopper()
//
// This is a convenience function.
func withTestExporters() Option {
	return func(c *Config) {
		WithTracesExporterEndpoint("localhost:4317")(c)
		WithTracesExporterInsecure(true)(c)
		WithMetricsExporterEndpoint("localhost:4317")(c)
		WithMetricsExporterInsecure(true)(c)
	}
}

type testErrorHandler struct {
}

func (t *testErrorHandler) Handle(err error) {
	fmt.Printf("test error handler handled error: %v\n", err)
}

// TODO REVIEW TEST - want default service name anyway
// func TestInvalidServiceName(t *testing.T) {
// 	logger := &testLogger{}
// 	shutdown, _ := ConfigureOpenTelemetry(WithLogger(logger))
// 	defer shutdown()

// 	expected := "invalid configuration: service name missing"
// 	logger.requireContains(t, expected)
// }

func testEndpointDisabled(t *testing.T, expected string, opts ...Option) {
	logger := &testLogger{}
	shutdown, err := ConfigureOpenTelemetry(
		append(opts,
			WithLogger(logger),
			WithServiceName("test-service"),
		)...,
	)
	require.NoError(t, err)
	defer shutdown()

	logger.requireContains(t, expected)
}

func TestTraceEndpointDisabled(t *testing.T) {
	testEndpointDisabled(
		t,
		expectedTracingDisabledMessage,
		WithTracesExporterEndpoint(""),
		WithExporterEndpoint(""),
	)
}

func TestMetricEndpointDisabled(t *testing.T) {
	testEndpointDisabled(
		t,
		expectedMetricsDisabledMessage,
		WithMetricsExporterEndpoint(""),
		WithExporterEndpoint(""),
	)
}

func testSignalDisabled(t *testing.T, expected string, opts ...Option) {
	logger := &testLogger{}
	shutdown, err := ConfigureOpenTelemetry(
		append(opts,
			WithLogger(logger),
			WithServiceName("test-service"),
		)...,
	)
	require.NoError(t, err)
	defer shutdown()

	logger.requireContains(t, expected)
}

func TestMetricsDisabled(t *testing.T) {
	testSignalDisabled(
		t,
		expectedMetricsDisabledConfigMessage,
		WithMetricsEnabled(false),
	)
}

func TestTracesDisabled(t *testing.T) {
	testSignalDisabled(
		t,
		expectedTracingDisabledConfigMessage,
		WithTracesEnabled(false),
	)
}

func TestValidConfig(t *testing.T) {
	logger := &testLogger{}

	// in order for tests to not have to timeout during
	// the shutdown call, we must direct them to a running
	// server, which means that it has to go to localhost:4317,
	// and it must be Insecure.
	stopper := dummyGRPCListener()
	defer stopper()

	shutdown, err := ConfigureOpenTelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		withTestExporters(),
	)
	require.NoError(t, err)
	defer shutdown()

	if len(logger.output) > 0 {
		t.Errorf("\nExpected: no logs\ngot: %v", logger.output)
	}
}

func TestInvalidEnvironment(t *testing.T) {
	setenv("OTEL_EXPORTER_OTLP_METRICS_INSECURE", "bleargh")
	defer unsetAllOtelEnvironmentVariables()

	logger := &testLogger{}

	_, err := ConfigureOpenTelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
	)
	require.ErrorContains(t, err, "environment error")
	logger.requireContains(t, "environment error")
}

func TestInvalidMetricsPushIntervalEnv(t *testing.T) {
	setenv("OTEL_EXPORTER_OTLP_METRICS_PERIOD", "300million")
	defer unsetAllOtelEnvironmentVariables()

	logger := &testLogger{}
	shutdown, err := ConfigureOpenTelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		withTestExporters(),
	)
	defer shutdown()
	assert.ErrorContains(t, err, "setup error: invalid metric reporting period")
}

func TestInvalidMetricsPushIntervalConfig(t *testing.T) {
	logger := &testLogger{}
	shutdown, err := ConfigureOpenTelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithMetricsReportingPeriod(-time.Second),
		withTestExporters(),
	)
	defer shutdown()

	assert.ErrorContains(t, err, "setup error: invalid metric reporting period")
}

func TestDebugEnabled(t *testing.T) {
	logger := &testLogger{}
	stopper := dummyGRPCListener()
	defer stopper()

	shutdown, _ := ConfigureOpenTelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		withTestExporters(),
		WithLogLevel("debug"),
		WithResourceAttributes(map[string]string{
			"attr1":     "val1",
			"host.name": "host456",
		}),
	)
	defer shutdown()
	output := strings.Join(logger.output[:], ",")
	assert.Contains(t, output, "debug logging enabled")
	assert.Contains(t, output, "test-service")
	assert.Contains(t, output, "localhost:4317")
	assert.Contains(t, output, "attr1")
	assert.Contains(t, output, "val1")
	assert.Contains(t, output, "host.name")
	assert.Contains(t, output, "host456")
}

func TestDefaultConfig(t *testing.T) {
	logger := &testLogger{}
	handler := &testErrorHandler{}
	config, err := newConfig(
		WithLogger(logger),
		WithErrorHandler(handler),
	)

	attributes := []attribute.KeyValue{
		attribute.String("host.name", host()),
		attribute.String("service.version", "unknown"),
		attribute.String("telemetry.sdk.name", "otelconfig"),
		attribute.String("telemetry.sdk.language", "go"),
		attribute.String("telemetry.sdk.version", version),
	}

	expected := &Config{
		ExporterEndpoint:                "localhost",
		ExporterEndpointInsecure:        false,
		TracesExporterEndpoint:          "",
		TracesExporterEndpointInsecure:  false,
		TracesEnabled:                   &trueVal,
		ServiceName:                     "",
		ServiceVersion:                  "unknown",
		MetricsExporterEndpoint:         "",
		MetricsExporterEndpointInsecure: false,
		MetricsEnabled:                  &trueVal,
		MetricsReportingPeriod:          "30s",
		LogLevel:                        "info",
		Headers:                         map[string]string{},
		TracesHeaders:                   map[string]string{},
		MetricsHeaders:                  map[string]string{},
		ResourceAttributes:              map[string]string{},
		Propagators:                     []string{"tracecontext", "baggage"},
		Resource:                        resource.NewWithAttributes(semconv.SchemaURL, attributes...),
		Logger:                          logger,
		ExporterProtocol:                "grpc",
		errorHandler:                    handler,
		Sampler:                         trace.AlwaysSample(),
	}
	assert.NoError(t, err)
	assert.Equal(t, expected, config)
}

func TestDefaultConfigMarshal(t *testing.T) {
	logger := &testLogger{}
	handler := &testErrorHandler{}
	config, err := newConfig(
		WithLogger(logger),
		WithErrorHandler(handler),
		WithShutdown(func(c *Config) error {
			return nil
		}),
	)
	assert.NoError(t, err)

	j, err := json.Marshal(config)
	assert.NoError(t, err)
	assert.NotEmpty(t, j)
}

func TestEnvironmentVariables(t *testing.T) {
	setEnvironment()
	defer unsetAllOtelEnvironmentVariables()

	logger := &testLogger{}
	handler := &testErrorHandler{}
	testConfig, err := newConfig(
		WithLogger(logger),
		WithErrorHandler(handler),
	)

	expectedHostname, err := os.Hostname()
	require.NoError(t, err)

	expectedConfiguredResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		attribute.String("host.name", expectedHostname),
		attribute.String("an.env.attr", "hi"),
		attribute.String("resource.clobber", "ENV_WON"),
		attribute.String("service.name", environmentOtelSettings["OTEL_SERVICE_NAME"]),
		attribute.String("service.version", environmentOtelSettings["OTEL_SERVICE_VERSION"]),
		attribute.String("telemetry.sdk.name", "otelconfig"),
		attribute.String("telemetry.sdk.language", "go"),
		attribute.String("telemetry.sdk.version", version),
	)

	expectedConfig := &Config{
		ServiceName:    environmentOtelSettings["OTEL_SERVICE_NAME"],
		ServiceVersion: environmentOtelSettings["OTEL_SERVICE_VERSION"],
		Resource:       expectedConfiguredResource,
		ResourceAttributes: map[string]string{
			"an.env.attr":      "hi",
			"resource.clobber": "ENV_WON",
		},
		Logger:                          logger,
		LogLevel:                        "debug",
		Propagators:                     []string{"b3", "w3c"},
		ExporterEndpoint:                environmentOtelSettings["OTEL_EXPORTER_OTLP_ENDPOINT"],
		ExporterEndpointInsecure:        true,
		ExporterProtocol:                Protocol(environmentOtelSettings["OTEL_EXPORTER_OTLP_PROTOCOL"]),
		Headers:                         map[string]string{"env-headers": "present", "header-clobber": "ENV_WON"},
		TracesEnabled:                   &trueVal,
		TracesExporterEndpoint:          environmentOtelSettings["OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"],
		TracesExporterEndpointInsecure:  true,
		TracesExporterProtocol:          Protocol(environmentOtelSettings["OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"]),
		TracesHeaders:                   map[string]string{"env-traces-headers": "present", "header-clobber": "ENV_WON"},
		MetricsEnabled:                  &falseVal,
		MetricsExporterEndpoint:         environmentOtelSettings["OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"],
		MetricsExporterEndpointInsecure: true,
		MetricsExporterProtocol:         Protocol(environmentOtelSettings["OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"]),
		MetricsHeaders:                  map[string]string{"env-metrics-headers": "present", "header-clobber": "ENV_WON"},
		MetricsReportingPeriod:          environmentOtelSettings["OTEL_EXPORTER_OTLP_METRICS_PERIOD"],
		Sampler:                         trace.AlwaysSample(),
		errorHandler:                    handler,
	}
	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, testConfig)
}

type testDetector struct{}

var _ resource.Detector = (*testDetector)(nil)

// Detect implements resource.Detector.
func (testDetector) Detect(ctx context.Context) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(attribute.String("a.test.detector", "detected")),
	)
}

func TestConfigurationOverrides(t *testing.T) {
	setEnvironment()
	defer unsetAllOtelEnvironmentVariables()

	logger := &testLogger{}
	handler := &testErrorHandler{}
	testConfig, err := newConfig(
		WithServiceName("service-name-from-code"),
		WithServiceVersion("service-version-from-code"),
		WithExporterEndpoint("https://an.endpoint.in.code"),
		WithExporterInsecure(false),
		WithTracesExporterEndpoint("traces-endpoint-from-code"),
		WithTracesExporterInsecure(false),
		WithMetricsExporterEndpoint("metrics-endpoint-from-code"),
		WithMetricsExporterInsecure(false),
		WithHeaders(map[string]string{"code-headers": "present", "header-clobber": "CODE_WON"}),
		WithTracesHeaders(map[string]string{"code-traces": "present", "header-clobber": "CODE_WON"}),
		WithMetricsHeaders(map[string]string{"code-metrics": "present", "header-clobber": "CODE_WON"}),
		WithLogLevel("info"),
		WithLogger(logger),
		WithErrorHandler(handler),
		WithPropagators([]string{"b3"}),
		WithExporterProtocol("http/json"),
		WithMetricsExporterProtocol("http/json"),
		WithTracesExporterProtocol("http/json"),
		WithResourceOption(resource.WithAttributes(
			attribute.String("a.code.attr", "hey"),
			attribute.String("resource.clobber", "CODE_WON"),
		)),
		WithResourceOption(resource.WithDetectors(&testDetector{})),
	)

	expectedConfiguredResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		attribute.String("a.code.attr", "hey"),
		attribute.String("a.test.detector", "detected"),
		attribute.String("an.env.attr", "hi"),
		attribute.String("resource.clobber", "ENV_WON"),
		attribute.String("host.name", host()),
		attribute.String("service.name", environmentOtelSettings["OTEL_SERVICE_NAME"]),
		attribute.String("service.version", environmentOtelSettings["OTEL_SERVICE_VERSION"]),
		attribute.String("telemetry.sdk.name", "otelconfig"),
		attribute.String("telemetry.sdk.language", "go"),
		attribute.String("telemetry.sdk.version", version),
	)

	expectedConfig := &Config{
		ServiceName:    environmentOtelSettings["OTEL_SERVICE_NAME"],
		ServiceVersion: environmentOtelSettings["OTEL_SERVICE_VERSION"],
		// what really matters is the configured resource, the result of all the merging of
		// defaults and detectors and overrides
		Resource: expectedConfiguredResource,
		// set on the otelconfig Config object, mostly via the OTEL_RESOURCE_ATTRIBUTES environment variable
		// we expect this struct property to match only the environment variable
		ResourceAttributes: map[string]string{
			"an.env.attr":      "hi",
			"resource.clobber": "ENV_WON",
		},
		// we expect these to match what was given in code code config.
		// Remember: the Resource property is what really matters as the result of merging all the
		// things and is what is actually used by the tracer/meter/etc.
		ResourceOptions: []resource.Option{
			resource.WithAttributes(
				attribute.String("a.code.attr", "hey"),
				attribute.String("resource.clobber", "CODE_WON"),
			),
			resource.WithDetectors(&testDetector{}),
		},
		Logger:                          logger,
		LogLevel:                        "debug",
		Propagators:                     []string{"b3", "w3c"},
		ExporterEndpoint:                environmentOtelSettings["OTEL_EXPORTER_OTLP_ENDPOINT"],
		ExporterEndpointInsecure:        true,
		ExporterProtocol:                Protocol(environmentOtelSettings["OTEL_EXPORTER_OTLP_PROTOCOL"]),
		Headers:                         map[string]string{"env-headers": "present", "header-clobber": "ENV_WON"},
		TracesEnabled:                   &trueVal,
		MetricsEnabled:                  &falseVal,
		TracesExporterEndpoint:          environmentOtelSettings["OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"],
		TracesExporterEndpointInsecure:  true,
		TracesExporterProtocol:          Protocol(environmentOtelSettings["OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"]),
		TracesHeaders:                   map[string]string{"env-traces-headers": "present", "header-clobber": "ENV_WON"},
		MetricsExporterEndpoint:         environmentOtelSettings["OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"],
		MetricsExporterEndpointInsecure: true,
		MetricsExporterProtocol:         Protocol(environmentOtelSettings["OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"]),
		MetricsHeaders:                  map[string]string{"env-metrics-headers": "present", "header-clobber": "ENV_WON"},
		MetricsReportingPeriod:          environmentOtelSettings["OTEL_EXPORTER_OTLP_METRICS_PERIOD"],
		Sampler:                         trace.AlwaysSample(),
		errorHandler:                    handler,
	}
	// Generic and signal-specific headers should merge
	expectedTraceHeaders := map[string]string{"env-headers": "present", "env-traces-headers": "present", "header-clobber": "ENV_WON"}
	expectedMetricsHeaders := map[string]string{"env-headers": "present", "env-metrics-headers": "present", "header-clobber": "ENV_WON"}

	assert.NoError(t, err)
	assert.Equal(t, expectedConfiguredResource, testConfig.Resource)
	assert.Equal(t, expectedConfig, testConfig)
	assert.Equal(t, expectedTraceHeaders, testConfig.getTracesHeaders())
	assert.Equal(t, expectedMetricsHeaders, testConfig.getMetricsHeaders())
}

type TestCarrier struct {
	values map[string]string
}

func (t TestCarrier) Keys() []string {
	keys := make([]string, 0, len(t.values))
	for k := range t.values {
		keys = append(keys, k)
	}
	return keys
}

func (t TestCarrier) Get(key string) string {
	return t.values[key]
}

func (t TestCarrier) Set(key string, value string) {
	t.values[key] = value
}

func TestConfigurePropagators1(t *testing.T) {
	stopper := dummyGRPCListener()
	defer stopper()

	mem1, _ := baggage.NewMember("keyone", "foo1")
	mem2, _ := baggage.NewMember("keytwo", "bar1")
	bag, _ := baggage.New(mem1, mem2)

	ctx := baggage.ContextWithBaggage(context.Background(), bag)

	logger := &testLogger{}
	shutdown, err := ConfigureOpenTelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		withTestExporters(),
	)
	assert.NoError(t, err)
	defer shutdown()

	ctx, finish := otel.Tracer("sampletracer").Start(ctx, "foo")
	defer finish.End()

	carrier := TestCarrier{values: map[string]string{}}
	prop := otel.GetTextMapPropagator()
	prop.Inject(ctx, carrier)
	baggage := carrier.Get("baggage")
	assert.Contains(t, baggage, "keyone=foo1")
	assert.Contains(t, baggage, "keytwo=bar1")
	assert.Greater(t, len(carrier.Get("traceparent")), 0)
}

func TestConfigurePropagators2(t *testing.T) {
	stopper := dummyGRPCListener()
	defer stopper()

	mem1, _ := baggage.NewMember("keyone", "foo1")
	mem2, _ := baggage.NewMember("keytwo", "bar1")
	bag, _ := baggage.New(mem1, mem2)

	ctx := baggage.ContextWithBaggage(context.Background(), bag)

	logger := &testLogger{}
	shutdown, err := ConfigureOpenTelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithPropagators([]string{"b3", "baggage", "tracecontext"}),
		withTestExporters(),
	)
	assert.NoError(t, err)
	defer shutdown()

	ctx, finish := otel.Tracer("sampletracer").Start(ctx, "foo")
	defer finish.End()

	carrier := TestCarrier{values: map[string]string{}}
	prop := otel.GetTextMapPropagator()
	prop.Inject(ctx, carrier)
	assert.Greater(t, len(carrier.Get("x-b3-traceid")), 0)
	baggage := carrier.Get("baggage")
	assert.Contains(t, baggage, "keyone=foo1")
	assert.Contains(t, baggage, "keytwo=bar1")
	assert.Greater(t, len(carrier.Get("traceparent")), 0)
}

func TestConfigurePropagators3(t *testing.T) {
	stopper := dummyGRPCListener()
	defer stopper()

	logger := &testLogger{}
	shutdown, err := ConfigureOpenTelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithPropagators([]string{"invalid"}),
		withTestExporters(),
	)
	defer shutdown()
	assert.ErrorContains(t, err, "invalid configuration: unsupported propagators. Supported options: b3,baggage,tracecontext,ottrace")
}

func host() string {
	host, _ := os.Hostname()
	return host
}

func TestConfigureResourcesAttributes(t *testing.T) {
	testCases := []struct {
		name               string
		codeConfig         Config
		envConfig          string
		expectedAttributes []attribute.KeyValue
	}{
		{
			name:       "default",
			codeConfig: Config{},
			envConfig:  "",
			expectedAttributes: []attribute.KeyValue{
				attribute.String("host.name", host()),
				attribute.String("telemetry.sdk.language", "go"),
				attribute.String("telemetry.sdk.name", "otelconfig"),
				attribute.String("telemetry.sdk.version", version),
			},
		},
		{
			name: "from code: ResourceAttributes",
			codeConfig: Config{
				ResourceAttributes: map[string]string{"label1": "value1", "label2": "value2"},
			},
			envConfig: "",
			expectedAttributes: []attribute.KeyValue{
				attribute.String("host.name", host()),
				attribute.String("label1", "value1"),
				attribute.String("label2", "value2"),
				attribute.String("telemetry.sdk.language", "go"),
				attribute.String("telemetry.sdk.name", "otelconfig"),
				attribute.String("telemetry.sdk.version", version),
			},
		},
		{
			name: "from code: ResourceOption",
			codeConfig: Config{
				ResourceOptions: []resource.Option{
					resource.WithAttributes(attribute.String("label1", "value1"), attribute.String("label2", "value2")),
				},
			},
			envConfig: "",
			expectedAttributes: []attribute.KeyValue{
				attribute.String("host.name", host()),
				attribute.String("label1", "value1"),
				attribute.String("label2", "value2"),
				attribute.String("telemetry.sdk.language", "go"),
				attribute.String("telemetry.sdk.name", "otelconfig"),
				attribute.String("telemetry.sdk.version", version),
			},
		},
		{
			name: "from code: ResourceOption and ResourceAttributes",
			codeConfig: Config{
				ResourceAttributes: map[string]string{
					"label1": "NOPE!",
					"label2": "Shouldn't see me.",
				},
				ResourceOptions: []resource.Option{
					resource.WithAttributes(
						attribute.String("label1", "I won!"),
						attribute.String("label2", "Horray!")),
				},
			},
			envConfig: "",
			expectedAttributes: []attribute.KeyValue{
				attribute.String("host.name", host()),
				attribute.String("label1", "I won!"),
				attribute.String("label2", "Horray!"),
				attribute.String("telemetry.sdk.language", "go"),
				attribute.String("telemetry.sdk.name", "otelconfig"),
				attribute.String("telemetry.sdk.version", version),
			},
		},
		{
			name:       "from env",
			codeConfig: Config{},
			envConfig:  "label1=value1,label2=value2",
			expectedAttributes: []attribute.KeyValue{
				attribute.String("host.name", host()),
				attribute.String("label1", "value1"),
				attribute.String("label2", "value2"),
				attribute.String("telemetry.sdk.language", "go"),
				attribute.String("telemetry.sdk.name", "otelconfig"),
				attribute.String("telemetry.sdk.version", version),
			},
		},
		{
			name: "from env: env beats code",
			codeConfig: Config{
				ResourceAttributes: map[string]string{
					"label1": "NOPE!",
					"label2": "Shouldn't see me.",
				},
				ResourceOptions: []resource.Option{
					resource.WithAttributes(
						attribute.String("label1", "NOPE!"),
						attribute.String("label2", "Shouldn't see me, either.")),
				},
			},
			envConfig: "host.name=hosty-mchostface,label1=ENV_WON,label2=ENV_WON,telemetry.sdk.language=ogg",
			expectedAttributes: []attribute.KeyValue{
				attribute.String("host.name", "hosty-mchostface"),
				attribute.String("label1", "ENV_WON"),
				attribute.String("label2", "ENV_WON"),
				attribute.String("telemetry.sdk.language", "ogg"),
				attribute.String("telemetry.sdk.name", "otelconfig"),
				attribute.String("telemetry.sdk.version", version),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setenv("OTEL_RESOURCE_ATTRIBUTES", tc.envConfig)
			defer unsetAllOtelEnvironmentVariables()

			resource, err := newResource(&tc.codeConfig)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedAttributes, resource.Attributes())
		})
	}
}

func TestServiceNameViaResourceAttributes(t *testing.T) {
	stopper := dummyGRPCListener()
	defer stopper()

	setenv("OTEL_RESOURCE_ATTRIBUTES", "service.name=test-service-b")
	defer unsetAllOtelEnvironmentVariables()

	logger := &testLogger{}
	shutdown, _ := ConfigureOpenTelemetry(
		WithLogger(logger),
		withTestExporters(),
	)
	defer shutdown()

	notExpected := "invalid configuration: service name missing"
	logger.requireNotContains(t, notExpected)
}

func TestEmptyResourceEnvVarsStillWin(t *testing.T) {
	stopper := dummyGRPCListener()
	defer stopper()

	setenv("OTEL_RESOURCE_ATTRIBUTES", "host.name=")
	defer unsetAllOtelEnvironmentVariables()

	shutdown, err := ConfigureOpenTelemetry(
		WithServiceName("test-service"),
		WithTracesExporterEndpoint("localhost:443"),
		WithResourceAttributes(map[string]string{
			"attr1":     "val1",
			"host.name": "",
		}),
		WithShutdown(func(c *Config) error {
			attrs := attribute.NewSet(c.Resource.Attributes()...)
			v, exists := attrs.Value("host.name")
			require.True(t, exists)
			assert.Equal(t, "", v.AsString(), "empty OTEL_RESOURCE_ATTRIBUTES still win")
			return nil
		}),
		withTestExporters(),
	)
	defer shutdown()
	assert.NoError(t, err)
}

func TestConfigWithResourceAttributes(t *testing.T) {
	stopper := dummyGRPCListener()
	defer stopper()

	shutdown, _ := ConfigureOpenTelemetry(
		WithServiceName("test-service"),
		WithResourceAttributes(map[string]string{
			"attr1": "val1",
			"attr2": "val2",
		}),
		WithShutdown(func(c *Config) error {
			attrs := attribute.NewSet(c.Resource.Attributes()...)
			v, ok := attrs.Value("attr1")
			assert.Equal(t, "val1", v.AsString())
			assert.True(t, ok)

			v, ok = attrs.Value("attr2")
			assert.Equal(t, "val2", v.AsString())
			assert.True(t, ok)
			return nil
		}),
		withTestExporters(),
	)
	defer shutdown()
}

func TestConfigWithResourceAttributesError(t *testing.T) {
	stopper := dummyGRPCListener()
	defer stopper()

	logger := &testLogger{}
	faultyResourceDetector := resource.StringDetector("", "", func() (string, error) {
		return "", errors.New("faulty resource detector")
	})

	_, err := ConfigureOpenTelemetry(
		WithLogger(logger),
		WithResourceAttributes(map[string]string{
			"attr1": "val1",
			"attr2": "val2",
		}),
		WithResourceOption(resource.WithDetectors(faultyResourceDetector)),
		WithShutdown(func(c *Config) error {
			attrs := attribute.NewSet(c.Resource.Attributes()...)
			v, ok := attrs.Value("attr1")
			assert.Equal(t, "val1", v.AsString())
			assert.True(t, ok)

			v, ok = attrs.Value("attr2")
			assert.Equal(t, "val2", v.AsString())
			assert.True(t, ok)

			logger.requireContains(t, "faulty resource detector")

			return nil
		}),
		withTestExporters(),
	)
	assert.ErrorContains(t, err, "faulty resource detector")
}

func TestConfigWithUnmergableResources(t *testing.T) {
	stopper := dummyGRPCListener()
	defer stopper()
	const otherSchemaURL = "https://test/other-schema-url"
	detect := resource.StringDetector(otherSchemaURL, "attr.key", func() (string, error) {
		return "attr.value", nil
	})

	_, err := ConfigureOpenTelemetry(
		WithServiceName("test-service"),
		WithResourceOption(resource.WithDetectors(detect)),
		withTestExporters(),
	)
	assert.ErrorContains(t, err, "conflicting Schema URL")
}

func TestThatEndpointsFallBackCorrectly(t *testing.T) {
	testCases := []struct {
		name            string
		configOpts      []Option
		tracesEndpoint  string
		tracesInsecure  bool
		metricsEndpoint string
		metricsInsecure bool
	}{
		{
			name:            "defaults",
			configOpts:      []Option{},
			tracesEndpoint:  "localhost:4317",
			tracesInsecure:  false,
			metricsEndpoint: "localhost:4317",
			metricsInsecure: false,
		},
		{
			name: "set generic endpoint / insecure",
			configOpts: []Option{
				WithExporterEndpoint("generic-url"),
				WithExporterInsecure(true),
			},
			tracesEndpoint:  "generic-url:4317",
			tracesInsecure:  true,
			metricsEndpoint: "generic-url:4317",
			metricsInsecure: true,
		},
		{
			name: "set specific endpoint / insecure",
			configOpts: []Option{WithExporterEndpoint("generic-url"),
				WithExporterInsecure(false),
				WithTracesExporterEndpoint("traces-url"),
				WithTracesExporterInsecure(true),
				WithMetricsExporterEndpoint("metrics-url:1234"),
				WithMetricsExporterInsecure(true),
			},
			tracesEndpoint:  "traces-url:4317",
			tracesInsecure:  true,
			metricsEndpoint: "metrics-url:1234",
			metricsInsecure: true,
		},
		{
			name: "set traces to protobuf, metrics default",
			configOpts: []Option{WithTracesExporterProtocol("http/protobuf"),
				WithTracesExporterEndpoint("traces-url"),
				WithTracesExporterInsecure(true),
			},
			tracesEndpoint:  "traces-url:4318",
			tracesInsecure:  true,
			metricsEndpoint: "localhost:4317",
			metricsInsecure: false,
		},
		{
			name: "set grpc endpoint with https scheme and no port, add port as helper",
			configOpts: []Option{WithExporterEndpoint("https://generic-url"),
				WithExporterProtocol("grpc"),
			},
			tracesEndpoint:  "generic-url:443",
			metricsEndpoint: "generic-url:443",
		},
		{
			name: "set grpc endpoint with https scheme and port, no update to port",
			configOpts: []Option{WithExporterEndpoint("https://generic-url:1234"),
				WithExporterProtocol("grpc"),
			},
			tracesEndpoint:  "generic-url:1234",
			metricsEndpoint: "generic-url:1234",
		},
		{
			name: "set grpc endpoint with http scheme and port, no update to port",
			configOpts: []Option{WithExporterEndpoint("http://generic-url:1234"),
				WithExporterProtocol("grpc"),
			},
			tracesEndpoint:  "generic-url:1234",
			metricsEndpoint: "generic-url:1234",
		},
		{
			name:            "defaults",
			configOpts:      []Option{},
			tracesEndpoint:  "localhost:4317",
			tracesInsecure:  false,
			metricsEndpoint: "localhost:4317",
			metricsInsecure: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := newConfig(tc.configOpts...)
			assert.NoError(t, err)
			tracesEndpoint, tracesInsecure := cfg.getTracesEndpoint()
			assert.Equal(t, tc.tracesEndpoint, tracesEndpoint)
			assert.Equal(t, tc.tracesInsecure, tracesInsecure)

			metricsEndpoint, metricsInsecure := cfg.getMetricsEndpoint()
			assert.Equal(t, tc.metricsEndpoint, metricsEndpoint)
			assert.Equal(t, tc.metricsInsecure, metricsInsecure)
		})
	}
}

func TestHttpProtoDefaultsToCorrectHostAndPort(t *testing.T) {
	logger := &testLogger{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debugf("received data from path: %s", r.URL)
	}))
	defer ts.Close()

	shutdown, err := ConfigureOpenTelemetry(
		WithLogger(logger),
		WithExporterEndpoint(ts.URL),
		WithExporterInsecure(true),
		WithExporterProtocol("http/protobuf"),
	)
	require.NoError(t, err)

	ctx := context.Background()
	tracer := otel.GetTracerProvider().Tracer("otelconfig-tests")
	_, span := tracer.Start(ctx, "test-span")
	span.End()
	shutdown()

	assert.True(t, len(logger.output) == 2)
	logger.requireContains(t, "received data from path: /v1/traces")
	logger.requireContains(t, "received data from path: /v1/metrics")
}

func TestCanConfigureCustomSampler(t *testing.T) {
	sampler := &testSampler{}
	config, err := newConfig(
		WithSampler(sampler),
	)

	assert.NoError(t, err)
	assert.Same(t, config.Sampler, sampler)
}

func TestCanUseCustomSampler(t *testing.T) {
	expectedSamplerProvidedAttribute := attribute.String("test", "value")
	sampler := &testSampler{
		decsision: trace.RecordAndSample,
		attributes: []attribute.KeyValue{
			expectedSamplerProvidedAttribute,
		},
	}

	traceServer := &dummyTraceServer{}
	stopper := dummyGRPCListenerWithTraceServer(traceServer)
	defer stopper()

	shutdown, err := ConfigureOpenTelemetry(
		WithSampler(sampler),
		withTestExporters(),
	)
	require.NoError(t, err)

	tracer := otel.GetTracerProvider().Tracer("otelconfig-tests")
	_, span := tracer.Start(context.Background(), "test-span")
	span.End()
	shutdown()

	spans := traceServer.recievedExportTraceServiceRequests[0].ResourceSpans[0].ScopeSpans[0].Spans
	require.Equal(t, 1, len(spans), "Should only be one span")

	attrs := spans[0].Attributes
	require.Equal(t, 1, len(attrs), "Should only be one attribute")

	attr := attrs[0]
	assert.Equal(t, string(expectedSamplerProvidedAttribute.Key), string(attr.Key))
	assert.Equal(t, expectedSamplerProvidedAttribute.Value.AsString(), attr.Value.GetStringValue())
}

func TestCanSetDefaultExporterEndpoint(t *testing.T) {
	DefaultExporterEndpoint = "http://custom.endpoint"
	config, err := newConfig()
	assert.NoError(t, err)
	assert.Equal(t, "http://custom.endpoint", config.ExporterEndpoint)
}

func TestCustomDefaultExporterEndpointDoesNotReplaceEnvVar(t *testing.T) {
	setEnvironment()
	defer unsetAllOtelEnvironmentVariables()

	DefaultExporterEndpoint = "http://custom.endpoint"
	config, err := newConfig()
	assert.NoError(t, err)
	assert.Equal(t, "http://generic-url", config.ExporterEndpoint)
}

func TestCustomDefaultExporterEndpointDoesNotReplaceOption(t *testing.T) {
	DefaultExporterEndpoint = "http://custom.endpoint"
	config, err := newConfig(
		WithExporterEndpoint("http://other.endpoint"),
	)
	assert.NoError(t, err)
	assert.Equal(t, "http://other.endpoint", config.ExporterEndpoint)
}

func TestSemanticConventionVersionMatchesUpstream(t *testing.T) {
	defaultResource := resource.Default()
	ourSchemaURL := semconv.SchemaURL
	assert.Equal(t, ourSchemaURL, defaultResource.SchemaURL())
}

func TestResourceDetectorsDontError(t *testing.T) {
	logger := &testLogger{}
	stopper := dummyGRPCListener()
	defer stopper()

	shutdown, err := ConfigureOpenTelemetry(
		WithLogger(logger),
		WithResourceOption(resource.WithHost()),
		withTestExporters(),
	)
	assert.NoError(t, err)
	defer shutdown()
}

func TestContribResourceDetectorsDontError(t *testing.T) {
	logger := &testLogger{}
	stopper := dummyGRPCListener()
	defer stopper()

	setenv("AWS_LAMBDA_FUNCTION_NAME", "lambdatest")
	defer os.Unsetenv("AWS_LAMBDA_FUNCTION_NAME")

	lambdaDetector := lambda.NewResourceDetector()

	_, err := ConfigureOpenTelemetry(
		WithLogger(logger),
		WithResourceOption(resource.WithDetectors(lambdaDetector)),
		withTestExporters(),
	)
	assert.NoError(t, err, "cannot merge resource due to conflicting Schema URL")
}

type testSampler struct {
	decsision  trace.SamplingDecision
	attributes []attribute.KeyValue
}

func (ts *testSampler) ShouldSample(parameters trace.SamplingParameters) trace.SamplingResult {
	return trace.SamplingResult{Decision: trace.RecordAndSample, Attributes: ts.attributes}
}

func (ts *testSampler) Description() string {
	return "testSampler"
}

// setenv is to stop the linter from complaining.
func setenv(key string, value string) {
	_ = os.Setenv(key, value)
}

// A map of the settings used to test configuring OpenTelemetry via environment variables.
var environmentOtelSettings = map[string]string{
	"OTEL_SERVICE_NAME":                   "test-service-name",
	"OTEL_SERVICE_VERSION":                "test-service-version",
	"OTEL_RESOURCE_ATTRIBUTES":            "an.env.attr=hi,resource.clobber=ENV_WON",
	"OTEL_LOG_LEVEL":                      "debug",
	"OTEL_PROPAGATORS":                    "b3,w3c",
	"OTEL_EXPORTER_OTLP_ENDPOINT":         "http://generic-url",
	"OTEL_EXPORTER_OTLP_INSECURE":         "true",
	"OTEL_EXPORTER_OTLP_HEADERS":          "env-headers=present,header-clobber=ENV_WON",
	"OTEL_EXPORTER_OTLP_PROTOCOL":         "http/protobuf",
	"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT":  "http://traces-url",
	"OTEL_EXPORTER_OTLP_TRACES_INSECURE":  "true",
	"OTEL_EXPORTER_OTLP_TRACES_HEADERS":   "env-traces-headers=present,header-clobber=ENV_WON",
	"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL":  "http/protobuf",
	"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT": "http://metrics-url",
	"OTEL_EXPORTER_OTLP_METRICS_INSECURE": "true",
	"OTEL_EXPORTER_OTLP_METRICS_HEADERS":  "env-metrics-headers=present,header-clobber=ENV_WON",
	"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL": "http/protobuf",
	"OTEL_EXPORTER_OTLP_METRICS_PERIOD":   "42s",
	"OTEL_METRICS_ENABLED":                "false",
}

// setEnvironment sets OTEL_ environment variables for testing config via environment.
func setEnvironment() {
	for varName, varValue := range environmentOtelSettings {
		setenv(varName, varValue)
	}
}

// Clears all OTEL_ environment variables, even ones not set by setEnvironment().
func unsetAllOtelEnvironmentVariables() {
	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, "OTEL_") {
			parts := strings.SplitN(envVar, "=", 2)
			_ = os.Unsetenv(parts[0])
		}
	}
}

func TestMain(m *testing.M) {
	unsetAllOtelEnvironmentVariables()
	os.Exit(m.Run())
}
