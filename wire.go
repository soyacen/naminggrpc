package pkg

import (
	"github.com/google/wire"
	dbx "github.com/soyacen/grocer/dbx"
	esx "github.com/soyacen/grocer/esx"
	goosex "github.com/soyacen/grocer/goosex"
	grpcx "github.com/soyacen/grocer/grpcx"
	kafkax "github.com/soyacen/grocer/kafkax"
	mongox "github.com/soyacen/grocer/mongox"
	nacosx "github.com/soyacen/grocer/nacosx"
	redisx "github.com/soyacen/grocer/redisx"
	registryx "github.com/soyacen/grocer/registryx"
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
	registryx.Provider,
)
