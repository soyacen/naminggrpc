package config

import temporalx "github.com/soyacen/grocer/grocer/temporalx"

type TemporalOptions struct {
	*temporalx.Options
}

func NewTemporalOptions(options *temporalx.Options) temporalx.IOptions {
	return &TemporalOptions{Options: options}
}

type TemporalConfig struct {
	*temporalx.Config
}

func (c *TemporalConfig) GetConfigs() map[string]temporalx.IOptions {
	configs := make(map[string]temporalx.IOptions)
	for k, v := range c.Config.GetConfigs() {
		configs[k] = NewTemporalOptions(v)
	}
	return configs
}

func NewTemporalConfig(config *temporalx.Config) temporalx.IConfig {
	return &TemporalConfig{Config: config}
}
