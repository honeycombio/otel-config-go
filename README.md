# otel-config-go (formerly otel-launcher-go)

This project is a configuration layer that chooses default values for configuration options that many OpenTelemetry users would ultimately configure manually, allowing for minimal code to quickly instrument with OpenTelemetry.

Latest release built with:

- OpenTelemetry Go [v1.27.0/v0.49.0](https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.27.0)
- OpenTelemetry Go Contrib [v1.27.0/v0.52.0](https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.27.0)
- OpenTelemetry Semantic Conventions [v1.25.0](https://github.com/open-telemetry/opentelemetry-go/tree/main/semconv/v1.25.0)

Minimum Go Version: `1.21`

See the OpenTelemetry SDK's [compatability matrix](https://github.com/open-telemetry/opentelemetry-go#compatibility) for more information.

## Getting started

```bash
go get github.com/honeycombio/otel-config-go
```

## Configure

Minimal setup - by default will send all telemetry via GRPC to `localhost:4317`

```go
import "github.com/honeycombio/otel-config-go/otelconfig"

func main() {
    otelShutdown, err := otelconfig.ConfigureOpenTelemetry()
    defer otelShutdown()
}
```

You can set headers directly instead.

```go
import "github.com/honeycombio/otel-config-go/otelconfig"

func main() {
    otelShutdown, err := otelconfig.ConfigureOpenTelemetry(
        otelconfig.WithServiceName("service-name"),
        otelconfig.WithHeaders(map[string]string{
            "service-auth-key": "value",
            "service-useful-field": "testing",
        }),
    )
    defer otelShutdown()
}
```

### Migrating from otel-launcher-go to otel-config-go

As of v1.8.0, this package has been renamed from `otel-launcher-go` to `otel-config-go`. When migrating to use the renamed package, all references to `launcher` should be changed to `otelconfig`.

## Configuration Options

| Config Option               | Env Variable                        | Required | Default              |
| --------------------------- | ----------------------------------- | -------- | -------------------- |
| WithServiceName             | OTEL_SERVICE_NAME                   | y        | -                    |
| WithServiceVersion          | OTEL_SERVICE_VERSION                | n        | -                    |
| WithHeaders                 | OTEL_EXPORTER_OTLP_HEADERS          | n        | {}                   |
| WithTracesHeaders           | OTEL_EXPORTER_OTLP_TRACES_HEADERS   | n        | {}                   |
| WithMetricsHeaders          | OTEL_EXPORTER_OTLP_METRICS_HEADERS  | n        | {}                   |
| WithExporterProtocol        | OTEL_EXPORTER_OTLP_PROTOCOL         | n        | grpc                 |
| WithTracesExporterEndpoint  | OTEL_EXPORTER_OTLP_TRACES_ENDPOINT  | n        | localhost:4317       |
| WithTracesExporterInsecure  | OTEL_EXPORTER_OTLP_TRACES_INSECURE  | n        | false                |
| WithMetricsExporterEndpoint | OTEL_EXPORTER_OTLP_METRICS_ENDPOINT | n        | localhost:4317       |
| WithMetricsExporterInsecure | OTEL_EXPORTER_OTLP_METRICS_INSECURE | n        | false                |
| WithLogLevel                | OTEL_LOG_LEVEL                      | n        | info                 |
| WithPropagators             | OTEL_PROPAGATORS                    | n        | tracecontext,baggage |
| WithResourceAttributes      | OTEL_RESOURCE_ATTRIBUTES            | n        | -                    |
| WithMetricsReportingPeriod  | OTEL_EXPORTER_OTLP_METRICS_PERIOD   | n        | 30s                  |
| WithMetricsEnabled          | OTEL_METRICS_ENABLED                | n        | true                 |
| WithTracesEnabled           | OTEL_TRACES_ENABLED                 | n        | true                 |

------

This is a joint effort alongside LightStep and is based their initial [otel-launcher-go](https://github.com/lightstep/otel-launcher-go). The intention is to contribute this to OpenTelemetry Go Contrib.
