package kafkax

import (
	"context"

	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	"github.com/soyacen/gox/conc/lazyload"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"kafkax",
	fx.Provide(NewReceivers, NewSenders),
	fx.Invoke(
		func(lc fx.Lifecycle, g *lazyload.Group[*kafka_sarama.Consumer]) {
			lc.Append(fx.StopHook(func(ctx context.Context) error {
				return g.Close(ctx)
			}))
		},
	),
	fx.Invoke(
		func(lc fx.Lifecycle, g *lazyload.Group[*kafka_sarama.Sender]) {
			lc.Append(fx.StopHook(func(ctx context.Context) error {
				return g.Close(ctx)
			}))
		},
	),
)
