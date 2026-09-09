package services

import (
	"strings"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Reason codes for the refusals this service raises. They are dotted constants rather than inline
// strings so a client can branch on them and a rename shows up at every use site at once.
const (
	ReasonMethodArchived        = "sales_fulfillment_method.archived"
	ReasonMethodNotFound        = "sales_fulfillment_method.not_found"
	ReasonMethodNotAllowed      = "sales_fulfillment_method.not_allowed_on_channel"
	ReasonMethodTypeUnsupported = "sales_fulfillment_method.type_not_implemented"
	ReasonMethodInUse           = "sales_fulfillment_method.in_use_as_default"
)

// SalesFulfillmentMethodDomainServiceImpl adds the archive lifecycle to the engine's default
// service. There is no suspend here, unlike a channel or a point: a method is either offered to new
// orders or it is not, and the fulfillments that already snapshotted it keep running either way, so
// a second "temporarily off" state would say nothing the archive flag does not.
type SalesFulfillmentMethodDomainServiceImpl struct {
	drif.DynamicResourceService
}

func NewSalesFulfillmentMethodDomainService(
	base drif.DynamicResourceService,
) *SalesFulfillmentMethodDomainServiceImpl {
	return &SalesFulfillmentMethodDomainServiceImpl{DynamicResourceService: base}
}

// Archive withdraws a method from new orders. Idempotent: archiving an archived method reports
// success.
//
// It refuses while any channel or sales point still names the method as its default, because the
// result would be a configuration that refuses every order it receives — the default resolves to a
// method no order is allowed to use, and the failure would surface at the till rather than here.
// Clearing the default first is the operator's decision to make explicitly.
func (this *SalesFulfillmentMethodDomainServiceImpl) Archive(
	ctx corectx.Context, methodId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.setArchived(ctx, methodId, true)
}

// Unarchive returns a method to the selectable set. It shares Archive's permission: it is the same
// power in reverse, so whoever may retire a policy may undo it.
func (this *SalesFulfillmentMethodDomainServiceImpl) Unarchive(
	ctx corectx.Context, methodId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.setArchived(ctx, methodId, false)
}

func (this *SalesFulfillmentMethodDomainServiceImpl) setArchived(
	ctx corectx.Context, methodId string, archived bool,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withTransaction(ctx, models.SalesFulfillmentMethodSchemaName,
		func(tranxCtx corectx.Context) error {
			method, err := loadRecord(tranxCtx, models.SalesFulfillmentMethodSchemaName,
				models.SalesFulfillmentMethodFieldId, methodId)
			if err != nil {
				return err
			}
			if method == nil {
				result = notFoundResult(models.SalesFulfillmentMethodSchemaName, methodId)
				return nil
			}
			if boolOf(method, basemodel.FieldIsArchived) == archived {
				result = mutateOk()
				return nil
			}

			if archived {
				inUse, err := this.isNamedAsDefault(tranxCtx, methodId)
				if err != nil {
					return err
				}
				if inUse {
					result = violationResult(models.SalesFulfillmentMethodSchemaName,
						ReasonMethodInUse,
						"this fulfillment method is still the default of a sales channel or sales "+
							"point; point them at another method before archiving it")
					return nil
				}
			}

			result = mutateOk()
			return writeChanges(tranxCtx, models.SalesFulfillmentMethodSchemaName, method,
				dmodel.DynamicFields{basemodel.FieldIsArchived: archived})
		})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// isNamedAsDefault reports whether any channel or point would be left pointing at this method.
// Both are checked because either one alone resolves an order's method, so leaving either behind
// produces the same broken configuration.
func (this *SalesFulfillmentMethodDomainServiceImpl) isNamedAsDefault(
	ctx corectx.Context, methodId string,
) (bool, error) {
	named, err := anyRecordWhere(ctx, models.SalesChannelSchemaName,
		models.SalesChannelFieldDefaultFulfillmentMethodId, methodId)
	if err != nil || named {
		return named, err
	}
	return anyRecordWhere(ctx, models.SalesPointSchemaName,
		models.SalesPointFieldDefaultFulfillmentMethodId, methodId)
}

// anyRecordWhere reports whether any row of a schema carries a value in one field. It asks for a
// single row rather than counting: the caller only needs to know that at least one exists, and
// fetching the rest to discard them would cost a table scan on a busy table.
func anyRecordWhere(
	ctx corectx.Context, schemaName string, field string, value string,
) (bool, error) {
	engine, err := engineFor(schemaName)
	if err != nil {
		return false, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(field, dmodel.Equals, value))

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return false, err
	}
	return found != nil && found.HasData && len(found.Data.Items) > 0, nil
}

// AssertAssignable is the one gate every path that attaches a method to new business goes through:
// order creation, and setting a channel or point default. It answers a refusal rather than an
// error, because each of these is something the caller can correct.
//
// channelId may be empty when the caller is not yet bound to a channel (setting a point default
// resolves the channel itself); the allowed-set check is then skipped rather than failing open on a
// channel nobody named.
func (this *SalesFulfillmentMethodDomainServiceImpl) AssertAssignable(
	ctx corectx.Context, methodId string, channelId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	method, err := loadRecord(ctx, models.SalesFulfillmentMethodSchemaName,
		models.SalesFulfillmentMethodFieldId, methodId)
	if err != nil {
		return nil, err
	}
	if method == nil {
		return violationResult(models.SalesFulfillmentMethodSchemaName,
			ReasonMethodNotFound,
			"no fulfillment method with id '"+methodId+"'"), nil
	}
	if boolOf(method, basemodel.FieldIsArchived) {
		return violationResult(models.SalesFulfillmentMethodSchemaName,
			ReasonMethodArchived,
			"an archived fulfillment method cannot be assigned to new business; the "+
				"fulfillments that already snapshotted it are unaffected"), nil
	}

	fulfillmentType := stringOf(method, models.SalesFulfillmentMethodFieldFulfillmentType)
	if models.FulfillmentType(fulfillmentType) != models.FulfillmentTypeKioskDispense {
		return violationResult(models.SalesFulfillmentMethodSchemaName,
			ReasonMethodTypeUnsupported,
			"fulfillment type '"+fulfillmentType+"' has no execution workflow yet; only "+
				string(models.FulfillmentTypeKioskDispense)+" can be fulfilled"), nil
	}

	if channelId != "" {
		allowed, err := this.IsAllowedOnChannel(ctx, channelId, methodId)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return violationResult(models.SalesFulfillmentMethodSchemaName,
				ReasonMethodNotAllowed,
				"this fulfillment method is not among those allowed on the sales channel"), nil
		}
	}

	return mutateOk(), nil
}

// IsAllowedOnChannel reports whether a channel has been configured to permit a method. Default-deny:
// a channel with no mapping rows permits nothing, so a newly created method is off everywhere until
// somebody allows it — the same rule the payment mapping beside it follows.
func (this *SalesFulfillmentMethodDomainServiceImpl) IsAllowedOnChannel(
	ctx corectx.Context, channelId string, methodId string,
) (bool, error) {
	if channelId == "" || methodId == "" {
		return false, nil
	}
	engine, err := engineFor(models.SalesChannelFulfillmentMethodSchemaName)
	if err != nil {
		return false, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.SalesChannelFulfillmentMethodFieldSalesChannelId:      channelId,
		models.SalesChannelFulfillmentMethodFieldFulfillmentMethodId: methodId,
	})
	if err != nil {
		return false, err
	}
	return found != nil && found.HasData, nil
}

// NormalizeMethodCode trims and upper-cases a method code before it is stored. Like a channel code
// it is immutable once written, so normalising afterwards would repoint every seed and caller that
// resolved the old spelling; it has to happen before the row exists.
func NormalizeMethodCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
