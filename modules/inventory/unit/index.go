package unit

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/app"
	repo "github.com/sky-as-code/nikki-erp/modules/inventory/unit/repository"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/transport"
)

func Init() error {
	err := errors.Join(
		repo.InitRepositories(),
		app.InitServices(),
		transport.InitTransport(),
	)

	return err
}
