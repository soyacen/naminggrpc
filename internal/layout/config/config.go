package config

import (
	context "context"
	"os"
	"strconv"

	"github.com/soyacen/gonfig/resource/nacos"
	"github.com/soyacen/gox/errorx"
	nacosx "github.com/soyacen/grocer/grocer/nacosx"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// RunEnv 获取运行环境 dev、test、pre、prod
func RunEnv() string {
	return os.Getenv("RUN_ENV")
}

func IsDev() bool {
	return RunEnv() == "dev"
}

func IsTest() bool {
	return RunEnv() == "test"
}

func IsPre() bool {
	return RunEnv() == "pre"
}

func IsProd() bool {
	return RunEnv() == "prod"
}

func LoadConfigFromNacos(ctx context.Context) error {
	if os.Getenv("NACOS_CONFIG_IP") == "" {
		return nil
	}
	configClient, err := nacosx.NewConfigClient(&nacosx.Options{
		IpAddr:              wrapperspb.String(os.Getenv("NACOS_CONFIG_IP")),
		Port:                wrapperspb.UInt64(errorx.Ignore(strconv.ParseUint(os.Getenv("NACOS_CONFIG_PORT"), 10, 64))),
		Namespace:           wrapperspb.String(os.Getenv("NACOS_CONFIG_NAMESPACE")),
		NotLoadCacheAtStart: wrapperspb.Bool(true),
	})
	if err != nil {
		return err
	}
	resource, err := nacos.New(configClient, os.Getenv("NACOS_CONFIG_GROUP"), os.Getenv("NACOS_CONFIG_DATA_ID"))
	if err != nil {
		return err
	}
	if err := LoadConfig(ctx, resource); err != nil {
		return err
	}
	return nil
}
