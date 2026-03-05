package cqrs_bus

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	identityCqrs "github.com/sky-as-code/nikki-erp/modules/drive/adapter/cqrs_bus/identity_cqrs"
)

func InitCqrsBusAdaper() error {
	err := errors.Join(
		deps.Register(identityCqrs.NewIdentityCqrsAdapter),
	)
	return err
}
