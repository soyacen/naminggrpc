package cronjob

import (
	"github.com/google/wire"
)

var Provider = wire.NewSet(NewService)
