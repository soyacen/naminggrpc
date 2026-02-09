package nacosgrpc

import (
	"context"

	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type DSN struct {
	clientParam     vo.NacosClientParam        // Nacos client parameters
	registerParam   vo.RegisterInstanceParam   // Registration instance parameters
	deregisterParam vo.DeregisterInstanceParam // Deregistration instance parameters
}

type DsnParser func(ctx context.Context, kind string, dsn string) (*DSN, error)

var DefaultDsnParser = ParseDsn

func ParseDsn(ctx context.Context, kind string, dsn string) (*DSN, error) {
	if kind == "registrar" {
		return parseRegistrarDSN(*u)
	}
	if kind == "resolver" {
		return parseResolverDSN(*u)
	}
	return nil, nil
}
