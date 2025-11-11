package unitcategory

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unitcategory/impl"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unitcategory/transport"
)

func InitSubModule() error {
	err := stdErr.Join(
		orm.RegisterEntity(impl.BuildUnitCategoryDescriptor()),
		deps.Register(impl.NewUnitCategoryEntRepository),
		deps.Register(impl.NewUnitCategoryServiceImpl),
		transport.InitTransport(),
	)

	return err
}
