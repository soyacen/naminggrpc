package otelx

import (
	"context"
	"os"
	"slices"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Attributes trace需要一些额外的信息
var Attributes []attribute.KeyValue

const (
	ServiceNameEnvKey      = "SERVICE_NAME"
	ServiceNamespaceEnvKey = "SERVICE_NAMESPACE"
	ServiceIDEnvKey        = "SERVICE_ID"
	ServiceVersionEnvKey   = "SERVICE_VERSION"
	ServiceNameKey         = "service.name"
	ServiceNamespaceKey    = "service.namespace"
	ServiceInstanceIDKey   = "service.instance.id"
	ServiceVersionKey      = "service.version"
)

func ServiceAttributes() []attribute.KeyValue {
	// 预分配最大可能容量的切片以避免多次重新分配
	attrs := make([]attribute.KeyValue, 0, 4)

	serviceName := os.Getenv(ServiceNameEnvKey)
	if serviceName != "" {
		attrs = append(attrs, attribute.Key(ServiceNameKey).String(serviceName))
	}

	serviceNamespace := os.Getenv(ServiceNamespaceEnvKey)
	if serviceNamespace != "" {
		attrs = append(attrs, attribute.Key(ServiceNamespaceKey).String(serviceNamespace))
	}

	serviceID := os.Getenv(ServiceIDEnvKey)
	if serviceID != "" {
		attrs = append(attrs, attribute.Key(ServiceInstanceIDKey).String(serviceID))
	}

	serviceVersion := os.Getenv(ServiceVersionEnvKey)
	if serviceVersion != "" {
		attrs = append(attrs, attribute.Key(ServiceVersionKey).String(serviceVersion))
	}

	return attrs
}

func NewShortResource() *resource.Resource {
	return resource.Default()
}

func NewResource(ctx context.Context) *resource.Resource {
	opts := []resource.Option{
		resource.WithAttributes(append(slices.Clone(Attributes), ServiceAttributes()...)...),
		resource.WithFromEnv(),
		resource.WithHost(),
		resource.WithHostID(),
		resource.WithTelemetrySDK(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithContainer(),
	}
	attributes, err := resource.New(ctx, opts...)
	if err != nil {
		return resource.Default()
	}
	return attributes
}
