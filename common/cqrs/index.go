package cqrs

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/common/logging"
	deps "github.com/sky-as-code/nikki-erp/common/util/deps_inject"
)

func InitSubModule() (modErr error) {
	err := deps.Provide(func(logger logging.LoggerService) CqrsBus {
		cqrsBus, err := NewWatermillCqrsBus(CqrsBusConfig{
			Logger: logger,
		})
		modErr = err
		return cqrsBus
	})

	modErr = errors.Join(modErr, err)
	return modErr
}
