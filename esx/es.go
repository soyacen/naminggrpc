package esx

import (
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/soyacen/goconc/lazyload"
	"github.com/spf13/cast"
)

func ConvertToUniversalOptions(options *Options) elasticsearch.Config {
	if options == nil {
		return elasticsearch.Config{}
	}

	opts := elasticsearch.Config{}

	// Addresses - A list of Elasticsearch nodes to use
	opts.Addresses = options.GetAddresses()

	// Username - Username for HTTP Basic Authentication
	opts.Username = options.GetUsername().GetValue()

	// Password - Password for HTTP Basic Authentication
	opts.Password = options.GetPassword().GetValue()

	// CloudID - Endpoint for the Elastic Service
	opts.CloudID = options.GetCloudId().GetValue()

	// APIKey - Base64-encoded token for authorization
	opts.APIKey = options.GetApiKey().GetValue()

	// ServiceToken - Service token for authorization
	opts.ServiceToken = options.GetServiceToken().GetValue()

	// CertificateFingerprint - SHA256 hex fingerprint
	opts.CertificateFingerprint = options.GetCertificateFingerprint().GetValue()

	// Header - Global HTTP request header
	if options.Header != nil {
		opts.Header = make(http.Header)
		for k, v := range options.GetHeader() {
			opts.Header[k] = []string{v}
		}
	}

	// CACert - PEM-encoded certificate authorities
	opts.CACert = options.GetCaCert().GetValue()

	// RetryOnStatus - List of status codes for retry
	opts.RetryOnStatus = cast.ToIntSlice(options.GetRetryOnStatus())

	// DisableRetry - Default: false
	opts.DisableRetry = options.GetDisableRetry().GetValue()

	// MaxRetries - Default: 3
	opts.MaxRetries = int(options.GetMaxRetries().GetValue())

	// CompressRequestBody - Default: false
	opts.CompressRequestBody = options.GetCompressRequestBody().GetValue()

	// CompressRequestBodyLevel - Default: gzip.DefaultCompression
	opts.CompressRequestBodyLevel = int(options.GetCompressRequestBodyLevel().GetValue())

	// PoolCompressor - If true, a sync.Pool based gzip writer is used. Default: false
	opts.PoolCompressor = options.GetPoolCompressor().GetValue()

	// DiscoverNodesOnStart - Discover nodes when initializing the client. Default: false
	opts.DiscoverNodesOnStart = options.GetDiscoverNodesOnStart().GetValue()

	// DiscoverNodesInterval - Discover nodes periodically. Default: disabled
	opts.DiscoverNodesInterval = time.Duration(options.GetDiscoverNodesInterval().GetValue())

	// EnableMetrics - Enable the metrics collection
	opts.EnableMetrics = options.GetEnableMetrics().GetValue()

	// EnableDebugLogger - Enable the debug logging
	opts.EnableDebugLogger = options.GetEnableDebugLogger().GetValue()

	// EnableCompatibilityMode - Enable sends compatibility header
	opts.EnableCompatibilityMode = options.GetEnableCompatibilityMode().GetValue()

	// DisableMetaHeader - Disable the additional "X-Elastic-Client-Meta" HTTP header
	opts.DisableMetaHeader = options.GetDisableMetaHeader().GetValue()

	return opts
}

func NewClients(config *Config) *lazyload.Group[*elasticsearch.Client] {
	return &lazyload.Group[*elasticsearch.Client]{
		New: func(key string) (*elasticsearch.Client, error) {
			configs := config.GetConfigs()
			options, ok := configs[key]
			if !ok {
				return nil, fmt.Errorf("es %s not found", key)
			}
			return NewClient(options)
		},
	}
}

func NewClient(options *Options) (*elasticsearch.Client, error) {
	return elasticsearch.NewClient(ConvertToUniversalOptions(options))
}
