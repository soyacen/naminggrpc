package redisx

import (
	"github.com/google/wire"
)

var Provider = wire.NewSet(
	NewRedisClients,
)
