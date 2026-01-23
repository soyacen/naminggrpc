package pkg

import (
	"github.com/google/wire"
	"github.com/soyacen/grocer/pkg/dbx"
	"github.com/soyacen/grocer/pkg/esx"
	"github.com/soyacen/grocer/pkg/goosex"
	"github.com/soyacen/grocer/pkg/grpcx"
	"github.com/soyacen/grocer/pkg/kafkax"
	"github.com/soyacen/grocer/pkg/mongox"
	"github.com/soyacen/grocer/pkg/nacosx"
	"github.com/soyacen/grocer/pkg/redisx"
)

var Provider = wire.NewSet(
	dbx.Provider,
	esx.Provider,
	goosex.Provider,
	grpcx.Provider,
	kafkax.Provider,
	mongox.Provider,
	nacosx.Provider,
	redisx.Provider,
)
