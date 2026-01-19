package grpcx

import (
	"math"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type ClientConfigs map[string]*ConsumerConfig

type ConsumerConfig struct {
	Target string `mapstructure:"target" json:"target" yaml:"target"`
}

// func NewClient() (*grpc.ClientConn, error) {
// 	return grpc.NewClient(
// 		"127.0.0.1:7777",
// 		DialogOptions()...,
// 	)
// }

func UnaryClientInterceptor(mdw ...grpc.UnaryClientInterceptor) grpc.DialOption {
	return grpc.WithChainUnaryInterceptor(append([]grpc.UnaryClientInterceptor{}, mdw...)...)
}

func StreamClientInterceptor(mdw ...grpc.StreamClientInterceptor) grpc.DialOption {
	return grpc.WithChainStreamInterceptor(append([]grpc.StreamClientInterceptor{}, mdw...)...)
}

func DialogOptions(opts ...grpc.DialOption) []grpc.DialOption {
	return append([]grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithWriteBufferSize(1024 * 1024),
		grpc.WithReadBufferSize(1024 * 1024),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(math.MaxInt32),
			grpc.MaxCallSendMsgSize(math.MaxInt32),
		),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
			Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // send pings even without active streams
		}),
	}, opts...)
}
