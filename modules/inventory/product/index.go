package product

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/inventory/product/app"
	repo "github.com/sky-as-code/nikki-erp/modules/inventory/product/repository"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/transport"
)

func Init() error {
	err := errors.Join(
		repo.InitRepositories(),
		app.InitServices(),
		transport.InitTransport(),
	)

	return err
}
