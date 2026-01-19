package pkg

import (
	"github.com/google/wire"
	"github.com/soyacen/grocer/pkg/pkg/dbx"
	"github.com/soyacen/grocer/pkg/pkg/esx"
	"github.com/soyacen/grocer/pkg/pkg/goosex"
	"github.com/soyacen/grocer/pkg/pkg/grpcx"
	"github.com/soyacen/grocer/pkg/pkg/kafkax"
	"github.com/soyacen/grocer/pkg/pkg/mongox"
	"github.com/soyacen/grocer/pkg/pkg/nacosx"
	"github.com/soyacen/grocer/pkg/pkg/redisx"
	"github.com/soyacen/grocer/pkg/pkg/registryx"
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
