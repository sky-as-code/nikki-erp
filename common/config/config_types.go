package config

import (
	"time"

	. "github.com/sky-as-code/nikki-erp/common/util/fault"
)

type ConfigName string

//go:generate mockgen -package mock -destination ../mock/config_loader_mock.go gitlab.cloudokyo.dev/backcommon/pkg/core/config/types ConfigLoader
type ConfigLoader interface {
	// Init must be called before using other methods in this interface
	// because it gives the strategy implementation a chance to load and cache
	// configuration values.
	Init() AppError
	Get(name string) (string, AppError)
}

//go:generate mockgen -package mock -destination ../mock/config_loader_mock.go gitlab.cloudokyo.dev/backcommon/pkg/core/config/types ConfigLoader
type ConfigService interface {
	// Init must be called before using other methods in this interface
	Init() AppError
	// Returns Git commit ID that this app is built
	GetAppVersion() string
	GetStr(name ConfigName, defaultVal ...interface{}) string
	GetStrArr(name ConfigName, defaultVal ...interface{}) []string
	GetDuration(configName ConfigName, defaultVal ...interface{}) time.Duration
	GetBool(configName ConfigName, defaultVal ...interface{}) bool
	GetUint(configName ConfigName, defaultVal ...interface{}) uint
	GetUint64(configName ConfigName, defaultVal ...interface{}) uint64
	GetInt(configName ConfigName, defaultVal ...interface{}) int
	GetInt32(configName ConfigName, defaultVal ...interface{}) int32
	GetInt64(configName ConfigName, defaultVal ...interface{}) int64
	GetFloat32(configName ConfigName, defaultVal ...interface{}) float32
}

type MapConfig struct {
	DbConnMaxLifeMinutes string `json:"DB_CONN_MAX_LIFE_MINUTES,omitempty"`
	DbHost               string `json:"DB_HOST,omitempty"`
	DbMaxOpenConns       string `json:"DB_MAX_OPEN_CONNS,omitempty"`
	DbMaxIdleConns       string `json:"DB_MAX_IDLE_CONNS,omitempty"`
	DbName               string `json:"DB_NAME,omitempty"`
	DbPassword           string `json:"DB_PASSWORD,omitempty"`
	DbPort               string `json:"DB_PORT,omitempty"`
	DbSslEnabled         string `json:"DB_SSL_ENABLED,omitempty"`
	DbUser               string `json:"DB_USER,omitempty"`
	//...//
}
