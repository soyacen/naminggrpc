package nacosgrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRegistrarDSN(t *testing.T) {
	tests := []struct {
		name        string
		rawURL      string
		expectedSvc string
		expectedGrp string
		expectedNs  string
		expectedIp  string
		expectedP   uint64
	}{
		{
			name:        "full registrar dsn",
			rawURL:      "nacos://127.0.0.1:8848/my-service?namespace=ns1&group=grp1&ip=1.2.3.4&port=9090&weight=20&ephemeral=false&cluster=c1&meta.key1=val1",
			expectedSvc: "my-service",
			expectedGrp: "grp1",
			expectedNs:  "ns1",
			expectedIp:  "1.2.3.4",
			expectedP:   9090,
		},
		{
			name:        "simple registrar dsn",
			rawURL:      "nacos://127.0.0.1:8848/my-service?ip=1.2.3.4&port=8080",
			expectedSvc: "my-service",
			expectedGrp: "DEFAULT_GROUP",
			expectedNs:  "public",
			expectedIp:  "1.2.3.4",
			expectedP:   8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseDsn(context.Background(), "registrar", tt.rawURL)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedSvc, parsed.RegisterParam.ServiceName)
			assert.Equal(t, tt.expectedGrp, parsed.RegisterParam.GroupName)
			assert.Equal(t, tt.expectedNs, parsed.ClientParam.ClientConfig.NamespaceId)
			assert.Equal(t, tt.expectedIp, parsed.RegisterParam.Ip)
			assert.Equal(t, tt.expectedP, parsed.RegisterParam.Port)

			if tt.name == "full registrar dsn" {
				assert.Equal(t, 20.0, parsed.RegisterParam.Weight)
				assert.Equal(t, false, parsed.RegisterParam.Ephemeral)
				assert.Equal(t, "c1", parsed.RegisterParam.ClusterName)
				assert.Equal(t, "val1", parsed.RegisterParam.Metadata["key1"])

				assert.Equal(t, "1.2.3.4", parsed.DeregisterParam.Ip)
				assert.Equal(t, uint64(9090), parsed.DeregisterParam.Port)
				assert.Equal(t, "c1", parsed.DeregisterParam.Cluster)
			}
		})
	}
}
