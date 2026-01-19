package dbx

import (
	"github.com/google/wire"
)

var Provider = wire.NewSet(
	NewDBs,
	NewSqlxDBs,
)
