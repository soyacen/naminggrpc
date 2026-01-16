package nacosx

import (
	"github.com/google/wire"
)

var Provider = wire.NewSet(
	NewConfigClients,
	NewRegistryClients,
)
