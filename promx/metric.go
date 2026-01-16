package otelx

import (
	"context"

	"github.com/soyacen/grocer/otelx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func SetPrometheusMeterProvider(ctx context.Context) {
	var opts []prometheus.Option
	exporter, err := prometheus.New(opts...)
	if err != nil {
		panic(err)
	}
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(otelx.NewResource(ctx)),
		sdkmetric.WithReader(exporter),
	)
	otel.SetMeterProvider(meterProvider)
}

func SetStdOutMeterProvider(ctx context.Context) {
	exporter, err := stdoutmetric.New()
	if err != nil {
		panic(err)
	}
	reader := sdkmetric.NewPeriodicReader(exporter)
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(meterProvider)
}
