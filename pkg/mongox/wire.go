package mongox

import (
	"github.com/google/wire"
)

var Provider = wire.NewSet(
	NewClients,
)
