package legacy

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/legacy/app"
	ext "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/legacy/infra/external"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/legacy/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/legacy/transport"
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
