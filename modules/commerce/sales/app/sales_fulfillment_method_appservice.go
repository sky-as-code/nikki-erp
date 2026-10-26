package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/catalog"
)

// NewSalesFulfillmentMethodApplicationService wraps the composable default with the archive
// lifecycle. The domain service is type-asserted rather than accepted as a parameter: the onion
// built it, and a mismatch is a wiring error worth failing loudly at boot instead of a nil call
// on the first request.
func NewSalesFulfillmentMethodApplicationService(
	base composable.CrudApplicationService,
) itCatalog.SalesFulfillmentMethodApplicationService {
	return &SalesFulfillmentMethodApplicationServiceImpl{
		CrudApplicationService: base,
		methodSvc:              base.DomainService().(itCatalog.SalesFulfillmentMethodDomainService),
	}
}

type SalesFulfillmentMethodApplicationServiceImpl struct {
	composable.CrudApplicationService
	methodSvc itCatalog.SalesFulfillmentMethodDomainService
}

// Archive and Unarchive share the built-in set_archived permission rather than inventing one:
// splitting them would let a role retire a method through one route while being refused on the
// other, which is the same power in two directions.
func (this *SalesFulfillmentMethodApplicationServiceImpl) Archive(
	ctx corectx.Context, cmd itCatalog.ArchiveFulfillmentMethodCommand,
) (*itCatalog.ArchiveFulfillmentMethodResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionSetArchived, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.methodSvc.Archive(ctx, readStringParam(cmd, paramRecordId))
}

func (this *SalesFulfillmentMethodApplicationServiceImpl) Unarchive(
	ctx corectx.Context, cmd itCatalog.ArchiveFulfillmentMethodCommand,
) (*itCatalog.ArchiveFulfillmentMethodResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionSetArchived, cmd); cErrs != nil || err != nil {
		return mutateFailure(cErrs, err)
	}
	return this.methodSvc.Unarchive(ctx, readStringParam(cmd, paramRecordId))
}
