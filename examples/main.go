package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/honeycombio/otel-config-go/otelconfig"
)

func main() {
	otelShutdown, otelFlush, err := otelconfig.ConfigureOpenTelemetry(
		otelconfig.WithResourceOption(
			resource.WithAttributes(
				attribute.String("resource.example_set_in_code", "CODE"),
				attribute.String("resource.example_clobber", "CODE_WON"),
			),
		),
	)

	if err != nil {
		log.Fatalf("error setting up OTel SDK - %e", err)
	}

	defer otelShutdown()
	tracer := otel.Tracer("my-app")
	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "doing-things")
	defer span.End()

	// optionally, flush any remaining spans
	otelFlush(context.Background())
}
