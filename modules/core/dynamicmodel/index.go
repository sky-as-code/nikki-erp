package dynamicmodel

import (
	_ "github.com/lib/pq"
	"go.uber.org/dig"
)

type InitParams struct {
	dig.In

	// Config config.ConfigService
	// Logger logging.LoggerService
}

func InitSubModule(params InitParams) error {
	// err := errors.Join(
	// 	deps.Invoke(registerCoreSearchPredicates),
	// )

	// return err
	return nil
}
