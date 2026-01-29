package grpcx

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/soyacen/gox/errorx"
)

type nacosOptions struct {
	ClusterName string
	Weight      float64
	GroupName   string
	Metadata    map[string]string
}

type NacosOption func(r *nacosOptions)

func ClusterName(clusterName string) NacosOption {
	return func(r *nacosOptions) {
		r.ClusterName = clusterName
	}
}

func Weight(weight float64) NacosOption {
	return func(r *nacosOptions) {
		r.Weight = weight
	}
}

func GroupName(name string) NacosOption {
	return func(r *nacosOptions) {
		r.GroupName = name
	}
}

type nacosRegistrar struct {
	ServiceName  string
	IP           string
	Port         int
	namingClient naming_client.INamingClient
	nacosOptions *nacosOptions
}

func (r *nacosRegistrar) Register(ctx context.Context) error {
	param := vo.RegisterInstanceParam{
		ServiceName: r.ServiceName,
		Ip:          r.IP,           // 服务实例IP
		Port:        uint64(r.Port), // 服务实例port
		ClusterName: r.nacosOptions.ClusterName,
		GroupName:   r.nacosOptions.GroupName,
		Weight:      r.nacosOptions.Weight,   // 权重
		Metadata:    r.nacosOptions.Metadata, // 扩展信息
		Enable:      true,                    // 是否上线
		Healthy:     true,                    // 是否健康
		Ephemeral:   true,                    // 是否临时实例
	}
	ok, err := r.namingClient.RegisterInstance(param)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("failed to register instance")
	}
	return nil
}

func (r *nacosRegistrar) Deregister(ctx context.Context) error {
	param := vo.DeregisterInstanceParam{
		Ip:          os.Getenv("POD_IP"),                                              // 服务实例IP
		Port:        errorx.Ignore(strconv.ParseUint(os.Getenv("GRPC_PORT"), 10, 64)), // 服务实例port
		Cluster:     r.nacosOptions.ClusterName,
		ServiceName: r.ServiceName,
		GroupName:   r.nacosOptions.GroupName,
		Ephemeral:   true, // 是否临时实例
	}
	ok, err := r.namingClient.DeregisterInstance(param)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("failed to deregister instance")
	}
	return nil
}

func NewNacosRegistrar(namingClient naming_client.INamingClient, serviceName string, port int, opts ...NacosOption) (Registrar, error) {
	r := &nacosRegistrar{
		ServiceName:  serviceName,
		IP:           "",
		Port:         port,
		namingClient: namingClient,
		nacosOptions: &nacosOptions{
			Weight: 10,
		},
	}
	for _, opt := range opts {
		opt(r.nacosOptions)
	}
	return r, nil
}
