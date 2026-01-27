//go:build tools

package tools

import (
	_ "github.com/soyacen/goose"
	_ "github.com/soyacen/gox"
	_ "github.com/soyacen/grocer/grocer"
	_ "github.com/ugorji/go/codec"
	_ "google.golang.org/genproto/googleapis/api"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	_ "google.golang.org/genproto/googleapis/rpc/status"
	_ "google.golang.org/grpc"
	_ "google.golang.org/protobuf/proto"
)
