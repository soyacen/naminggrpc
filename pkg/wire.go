package pkg

import (
	"github.com/google/wire"
	dbx "github.com/soyacen/grocer/pkg/dbx"
	esx "github.com/soyacen/grocer/pkg/esx"
	goosex "github.com/soyacen/grocer/pkg/goosex"
	grpcx "github.com/soyacen/grocer/pkg/grpcx"
	kafkax "github.com/soyacen/grocer/pkg/kafkax"
	mongox "github.com/soyacen/grocer/pkg/mongox"
	nacosx "github.com/soyacen/grocer/pkg/nacosx"
	redisx "github.com/soyacen/grocer/pkg/redisx"
	registryx "github.com/soyacen/grocer/pkg/registryx"
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
