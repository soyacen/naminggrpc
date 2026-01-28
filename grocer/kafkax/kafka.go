package kafkax

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	"github.com/soyacen/gox/conc/lazyload"
)

func (options *Options) ToConfig() *sarama.Config {
	saramaConfig := options.ToConfig()
	if options.GetVersion() != nil {
		version, err := sarama.ParseKafkaVersion(options.GetVersion().GetValue())
		if err != nil {
			panic(err)
		}
		saramaConfig.Version = version
	}
	return saramaConfig
}

func (options *Options) ToSenderOptionFuncs() []kafka_sarama.SenderOptionFunc {
	return []kafka_sarama.SenderOptionFunc{}
}

func NewReceiver(options *Options) (*kafka_sarama.Consumer, error) {
	return kafka_sarama.NewConsumer(
		options.GetBrokers(),
		options.ToConfig(),
		options.GetReceiver().GetGroupId().GetValue(),
		options.GetTopic().GetValue(),
	)
}

func NewSender(options *Options) (*kafka_sarama.Sender, error) {
	return kafka_sarama.NewSender(
		options.GetBrokers(),
		options.ToConfig(),
		options.GetTopic().GetValue(),
		options.ToSenderOptionFuncs()...,
	)
}

func NewReceivers(config *Config) *lazyload.Group[*kafka_sarama.Consumer] {
	return &lazyload.Group[*kafka_sarama.Consumer]{
		New: func(key string) (*kafka_sarama.Consumer, error) {
			options, ok := config.GetConfigs()[key]
			if !ok {
				return nil, fmt.Errorf("kafka %s not found", key)
			}
			return NewReceiver(options)
		},
		Finalize: func(ctx context.Context, consumer *kafka_sarama.Consumer) error {
			return consumer.Close(ctx)
		},
	}
}

func NewSenders(config *Config) *lazyload.Group[*kafka_sarama.Sender] {
	return &lazyload.Group[*kafka_sarama.Sender]{
		New: func(key string) (*kafka_sarama.Sender, error) {
			options, ok := config.GetConfigs()[key]
			if !ok {
				return nil, fmt.Errorf("kafka %s not found", key)
			}
			return NewSender(options)
		},
		Finalize: func(ctx context.Context, obj *kafka_sarama.Sender) error {
			return obj.Close(ctx)
		},
	}
}
