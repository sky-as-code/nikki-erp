package repository

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitRepositories() error {
	err := errors.Join(
		deps.Register(NewPartyEntRepository),
		deps.Register(NewCommChannelEntRepository),
		deps.Register(NewRelationshipEntRepository),
	)
	return err
}
