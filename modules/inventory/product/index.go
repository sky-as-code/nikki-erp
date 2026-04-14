package product

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/inventory/product/app"
	ext "github.com/sky-as-code/nikki-erp/modules/inventory/product/infra/external"
	repo "github.com/sky-as-code/nikki-erp/modules/inventory/product/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/transport"
)

func Init() error {
	err := errors.Join(
		ext.InitExternal(),
		repo.InitRepositories(),
		app.InitServices(),
		transport.InitTransport(),
	)

	return err
}
