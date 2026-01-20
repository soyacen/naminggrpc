package jeagerx

import (
	"context"

	"github.com/soyacen/grocer/pkg/otelx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

func (r *RetryConfig) ToGrpcRetryConfig() otlptracegrpc.RetryConfig {
	if r == nil {
		return otlptracegrpc.RetryConfig{}
	}
	return otlptracegrpc.RetryConfig{
		Enabled:         r.GetEnabled().GetValue(),
		InitialInterval: r.GetInitialInterval().AsDuration(),
		MaxElapsedTime:  r.GetMaxElapsedTime().AsDuration(),
		MaxInterval:     r.GetMaxInterval().AsDuration(),
	}
}

func (r *RetryConfig) ToHttpRetryConfig() otlptracehttp.RetryConfig {
	return otlptracehttp.RetryConfig{
		Enabled:         r.GetEnabled().GetValue(),
		InitialInterval: r.GetInitialInterval().AsDuration(),
		MaxElapsedTime:  r.GetMaxElapsedTime().AsDuration(),
		MaxInterval:     r.GetMaxInterval().AsDuration(),
	}
}

func SetStdOutTracerProvider(ctx context.Context, config *Config) {
	exporter, err := stdouttrace.New()
	if err != nil {
		panic(err)
	}
	setTracerProvider(ctx, config, exporter)
}

func SetHttpTracerProvider(ctx context.Context, config *Config) {
	var opts []otlptracehttp.Option
	httpOptions := config.GetHttpOptions()
	if httpOptions.GetEndpoint() != nil {
		opts = append(opts, otlptracehttp.WithEndpoint(httpOptions.GetEndpoint().GetValue()))
	}
	if httpOptions.GetEndpointUrl() != nil {
		opts = append(opts, otlptracehttp.WithEndpointURL(httpOptions.GetEndpointUrl().GetValue()))
	}
	if httpOptions.GetCompression() != nil {
		opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.Compression(httpOptions.GetCompression().GetValue())))
	}
	if httpOptions.GetUrlPath() != nil {
		opts = append(opts, otlptracehttp.WithURLPath(httpOptions.GetUrlPath().GetValue()))
	}
	if httpOptions.GetTlsConfig() != nil {
		opts = append(opts, otlptracehttp.WithTLSClientConfig(httpOptions.GetTlsConfig().AsConfig()))
	}
	if httpOptions.GetInsecure() != nil {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if len(httpOptions.GetHeaders()) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(httpOptions.GetHeaders()))
	}
	if httpOptions.GetTimeout() != nil {
		opts = append(opts, otlptracehttp.WithTimeout(httpOptions.GetTimeout().AsDuration()))
	}
	if httpOptions.GetRetry() != nil {
		opts = append(opts, otlptracehttp.WithRetry(httpOptions.GetRetry().ToHttpRetryConfig()))
	}
	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		panic(err)
	}
	setTracerProvider(ctx, config, exporter)
}

func SetGrpcTraceExporter(ctx context.Context, config *Config) {
	var grpcOpts []otlptracegrpc.Option
	grpcOptions := config.GetGrpcOptions()
	if grpcOptions.GetEndpoint() != nil {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithEndpoint(grpcOptions.GetEndpoint().GetValue()))
	}
	if grpcOptions.GetEndpointUrl() != nil {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithEndpointURL(grpcOptions.GetEndpointUrl().GetValue()))
	}
	if grpcOptions.GetEndpoint() != nil {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithEndpoint(grpcOptions.GetEndpoint().GetValue()))
	}
	if grpcOptions.GetInsecure() != nil {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithInsecure())
	}
	if grpcOptions.GetReconnectionPeriod() != nil {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithReconnectionPeriod(grpcOptions.GetReconnectionPeriod().AsDuration()))
	}
	if grpcOptions.GetCompressor() != nil {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithCompressor(grpcOptions.GetCompressor().GetValue()))
	}
	if len(grpcOptions.GetHeaders()) > 0 {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithHeaders(grpcOptions.GetHeaders()))
	}
	if grpcOptions.GetTlsConfig() != nil {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(grpcOptions.GetTlsConfig().AsConfig())))
	}
	if grpcOptions.GetServiceConfig() != nil {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithServiceConfig(grpcOptions.GetServiceConfig().GetValue()))
	}
	if grpcOptions.GetTimeout() != nil {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithTimeout(grpcOptions.GetTimeout().AsDuration()))
	}

	if grpcOptions.GetRetry() != nil {
		grpcOpts = append(grpcOpts, otlptracegrpc.WithRetry(grpcOptions.GetRetry().ToGrpcRetryConfig()))
	}

	exporter, err := otlptracegrpc.New(ctx, grpcOpts...)
	if err != nil {
		panic(err)
	}
	setTracerProvider(ctx, config, exporter)
}

func setTracerProvider(ctx context.Context, o *Config, exporter sdktrace.SpanExporter) {
	var bcOpts []sdktrace.BatchSpanProcessorOption
	tpOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exporter, bcOpts...),
		sdktrace.WithResource(otelx.NewShortResource()),
		sdktrace.WithSampler(newSampler(o.GetSamplingRate().GetValue())),
	}
	tracerProvider := sdktrace.NewTracerProvider(tpOpts...)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{}))
}

func newSampler(samplingRate float64) sdktrace.Sampler {
	var sampler sdktrace.Sampler
	switch {
	case samplingRate >= 1:
		sampler = sdktrace.AlwaysSample()
	case samplingRate < 0:
		sampler = sdktrace.NeverSample()
	default:
		sampler = sdktrace.TraceIDRatioBased(samplingRate)
	}
	return sampler
}
