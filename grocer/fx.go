package grocer

import (
	"github.com/soyacen/grocer/grocer/dbx"
	"github.com/soyacen/grocer/grocer/esx"
	"github.com/soyacen/grocer/grocer/kafkax"
	"github.com/soyacen/grocer/grocer/mongox"
	"github.com/soyacen/grocer/grocer/nacosx"
	"github.com/soyacen/grocer/grocer/redisx"
	"github.com/soyacen/grocer/grocer/s3x"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"grocer",
	dbx.Module,
	esx.Module,
	kafkax.Module,
	mongox.Module,
	nacosx.Module,
	redisx.Module,
	s3x.Module,
)
