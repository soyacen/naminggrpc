package config

import "go.uber.org/fx"

var Module = fx.Module(
	"config",
	fx.Provide(GetConfig, GetDb, GetEs, GetJeager, GetKafka, GetMongo, GetNacos, GetPyroscope, GetRedis, GetS3),
)
