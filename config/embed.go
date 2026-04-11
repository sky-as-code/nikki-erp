package config

import (
	_ "embed"

	mod "github.com/sky-as-code/nikki-erp/modules"
)

//go:embed config.default.yaml
var DefaultConfigYaml []byte

func GetDefaultConfigYaml() mod.DefaultConfig {
	return mod.DefaultConfig(DefaultConfigYaml)
}
