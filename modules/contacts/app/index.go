package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitServices() error {
	err := errors.Join(
		deps.Register(NewContactsEnumServiceImpl),
		deps.Register(NewPartyServiceImpl),
		deps.Register(NewCommChannelServiceImpl),
		deps.Register(NewRelationshipServiceImpl),
	)
	return err
}
