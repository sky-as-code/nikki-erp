package loader

//go:generate mockgen -package mock -destination ../mock/config_loader_mock.go gitlab.cloudokyo.dev/backcommon/pkg/core/config/types ConfigLoader
type ConfigLoader interface {
	// Init must be called before using other methods in this interface
	// because it gives the strategy implementation a chance to load and cache
	// configuration values.
	Init() error
	Get(name string) (string, error)
}
