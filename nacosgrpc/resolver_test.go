package nacosgrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNacosResolver_ServiceNameExtraction(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		expected string
	}{
		{
			name:     "simple path",
			rawURL:   "nacos://127.0.0.1:8848/my-service",
			expected: "my-service",
		},
		{
			name:     "path with leading slash",
			rawURL:   "nacos://127.0.0.1:8848//my-service",
			expected: "my-service",
		},
		{
			name:     "path with parameters",
			rawURL:   "nacos://127.0.0.1:8848/my-service?group=prod",
			expected: "my-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseDsn(context.Background(), "resolver", tt.rawURL)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, parsed.SubscribeParam.ServiceName)
		})
	}
}

func TestParseDSN(t *testing.T) {
	tests := []struct {
		name        string
		rawURL      string
		expectedSvc string
		expectedGrp string
		expectedNs  string
	}{
		{
			name:        "full dsn",
			rawURL:      "nacos://user:pass@127.0.0.1:8848/my-service?namespace=ns1&group=grp1&clusters=c1,c2&timeout=5000",
			expectedSvc: "my-service",
			expectedGrp: "grp1",
			expectedNs:  "ns1",
		},
		{
			name:        "simple dsn",
			rawURL:      "nacos://127.0.0.1:8848/my-service",
			expectedSvc: "my-service",
			expectedGrp: "DEFAULT_GROUP",
			expectedNs:  "public",
		},
		{
			name:        "dsn with alias params",
			rawURL:      "nacos://127.0.0.1:8848/my-service?namespace=ns2&group=grp2",
			expectedSvc: "my-service",
			expectedGrp: "grp2",
			expectedNs:  "ns2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseDsn(context.Background(), "resolver", tt.rawURL)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedSvc, parsed.SubscribeParam.ServiceName)
			assert.Equal(t, tt.expectedGrp, parsed.SubscribeParam.GroupName)
			assert.Equal(t, tt.expectedNs, parsed.ClientParam.ClientConfig.NamespaceId)

			if tt.name == "full dsn" {
				assert.Equal(t, []string{"c1", "c2"}, parsed.SubscribeParam.Clusters)
				assert.Equal(t, uint64(5000), parsed.ClientParam.ClientConfig.TimeoutMs)
				assert.Equal(t, "user", parsed.ClientParam.ClientConfig.Username)
				assert.Equal(t, "pass", parsed.ClientParam.ClientConfig.Password)
			}
		})
	}
}
