package app

import (
	"strings"

	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	orgDefaults "github.com/sky-as-code/nikki-erp/modules/commerce/infra/sales/defaults"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle/defaults"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

const (
	defaultCurrencyCode = "VND"
	paymentMethodAlias  = "payment_method_"
)

// writeOrder lists the defaulted resources parents first. The association rows between them are
// written after, through their owners' services.
var writeOrder = []string{
	models.SalesChannelSchemaName, models.SalesFulfillmentMethodSchemaName, models.SalesPointSchemaName,
	models.SalesPricelistSchemaName,
}

// cleanUpOrder lists every org-scoped resource children first, as their foreign keys require. The
// association rows are not here: they have no resource engine and are removed before these.
var cleanUpOrder = []string{
	models.SalesBillLineSchemaName, models.SalesBillRelationSchemaName,
	models.SalesBillingInstructionSchemaName, models.SalesBillingIssuanceAttemptSchemaName,
	models.SalesRefundPaymentSchemaName, models.SalesBillSchemaName,
	models.SalesFulfillmentAttemptItemSchemaName, models.SalesFulfillmentAttemptSchemaName,
	models.SalesOrderFulfillmentItemSchemaName, models.SalesFulfillmentTargetChangeSchemaName,
	models.SalesOrderFulfillmentSchemaName, models.SalesReturnSchemaName, models.SalesReturnLineSchemaName,
	models.SalesOrderSchemaName, models.SalesPricelistSchemaName, models.SalesPointSchemaName,
	models.SalesQuotationSchemaName, models.SalesChannelSchemaName, models.SalesComboComponentSchemaName,
	models.SalesComboSchemaName, models.SalesFiscalRequestSchemaName, models.SalesFulfillmentMethodSchemaName,
	models.SalesFulfillmentRequestLineSchemaName, models.SalesFulfillmentRequestSchemaName,
	models.SalesIntegrationOutboxSchemaName, models.SalesManualDiscountSchemaName,
	models.SalesOrderAdjustmentSchemaName, models.SalesOrderEventSchemaName,
	models.SalesOrderLineAllocationSchemaName, models.SalesOrderLineComponentSchemaName,
	models.SalesOrderLineSchemaName, models.SalesPaymentSchemaName,
	models.SalesPricelistItemSchemaName, models.SalesPromotionCompatibilitySchemaName,
	models.SalesPromotionConditionGroupSchemaName, models.SalesPromotionConditionTargetSchemaName,
	models.SalesPromotionConditionSchemaName, models.SalesPromotionProgramSchemaName,
	models.SalesPromotionRewardSchemaName, models.SalesQuotationLineSchemaName,
	models.SalesVoucherCodeSchemaName, models.SalesVoucherRedemptionSchemaName,
}

type OrgLifecycleParam struct {
	dig.In

	Lookup         itExt.OrgDefaultsLookupExtService
	PaymentMethods itExt.PaymentMethodExtService
	ChannelPayment *services.ChannelPaymentDomainServiceImpl
	PointPayment   *services.PointPaymentDomainServiceImpl
	Sequences      itExt.SequenceExtService
	SettingsSeed   itExt.OrgSettingsSeedExtService
}

func NewOrgLifecycleAppService(param OrgLifecycleParam) orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{param: param, cleanup: dynamicresource.NewOrgDataCleanup(cleanUpOrder...)}
}

// OrgLifecycleAppServiceImpl gives a new org the system sales channels, their payment and
// fulfillment options, a back-office sales point, a default price list and its document series.
type OrgLifecycleAppServiceImpl struct {
	param   OrgLifecycleParam
	cleanup dynamicresource.OrgDataCleanup
}

func (this *OrgLifecycleAppServiceImpl) InitOrg(ctx corectx.Context, orgId model.Id) error {
	svcs, err := dynamicresource.ResolveDomainServices(writeOrder...)
	if err != nil {
		return err
	}
	return dynamicresource.InTransaction(ctx, svcs[0].Repository(), func(tranxCtx corectx.Context) error {
		aliases, err := this.externalAliases(tranxCtx)
		if err != nil {
			return err
		}
		for _, svc := range svcs {
			rows, err := defaults.Load(orgDefaults.Files, svc.Schema().Name())
			if err != nil {
				return err
			}
			if err := dynamicresource.CreateDefaults(tranxCtx, svc, rows, aliases, orgId); err != nil {
				return err
			}
		}
		if err := this.createAssociations(tranxCtx, aliases, orgId); err != nil {
			return err
		}
		return this.registerSequences(tranxCtx, orgId)
	})
}

func (this *OrgLifecycleAppServiceImpl) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	svcs, err := dynamicresource.ResolveDomainServices(models.SalesChannelSchemaName)
	if err != nil {
		return err
	}
	return dynamicresource.InTransaction(ctx, svcs[0].Repository(), func(tranxCtx corectx.Context) error {
		if _, err := this.param.ChannelPayment.DeleteByOrg(tranxCtx, orgId); err != nil {
			return err
		}
		if _, err := this.param.PointPayment.DeleteByOrg(tranxCtx, orgId); err != nil {
			return err
		}
		if err := services.DeleteOrgChannelFulfillmentMethods(tranxCtx, orgId); err != nil {
			return err
		}
		return this.cleanup.CleanUpOrg(tranxCtx, orgId)
	})
}

// createAssociations maps payment and fulfillment methods onto the channels just created.
func (this *OrgLifecycleAppServiceImpl) createAssociations(
	ctx corectx.Context, aliases *defaults.Aliases, orgId model.Id,
) error {
	payments, err := this.loadResolved(models.SalesChannelPaymentRelSchemaName, aliases, orgId)
	if err != nil {
		return err
	}
	for _, row := range withKnownPaymentMethods(payments, aliases) {
		err := this.param.ChannelPayment.EnableInOrg(ctx, orgId,
			stringField(row, models.SalesChannelPaymentRelFieldSalesChannelId),
			stringField(row, models.SalesChannelPaymentRelFieldPaymentMethodId))
		if err != nil {
			return err
		}
	}

	fulfillments, err := this.loadResolved(models.SalesChannelFulfillmentMethodSchemaName, aliases, orgId)
	if err != nil {
		return err
	}
	for _, row := range fulfillments {
		err := services.AllowFulfillmentMethodOnChannel(ctx, orgId,
			stringField(row, models.SalesChannelFulfillmentMethodFieldSalesChannelId),
			stringField(row, models.SalesChannelFulfillmentMethodFieldFulfillmentMethodId))
		if err != nil {
			return err
		}
	}
	return nil
}

// loadResolved reads an association file with its aliases resolved. A payment method this tenant
// lacks resolves to nothing and is left as its "@alias", which withKnownPaymentMethods drops.
func (this *OrgLifecycleAppServiceImpl) loadResolved(
	schemaName string, aliases *defaults.Aliases, orgId model.Id,
) ([]dmodel.DynamicFields, error) {
	rows, err := defaults.Load(orgDefaults.Files, schemaName)
	if err != nil {
		return nil, err
	}
	resolved := make([]dmodel.DynamicFields, 0, len(rows))
	for _, row := range withKnownPaymentMethods(rows, aliases) {
		out, _, err := aliases.Resolve(row, orgId)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, out)
	}
	return resolved, nil
}

func stringField(row dmodel.DynamicFields, field string) string {
	value, _ := row[field].(string)
	return value
}

// externalAliases names what the defaults reference outside Sales: the currency, and the
// tenant's payment methods by code.
func (this *OrgLifecycleAppServiceImpl) externalAliases(ctx corectx.Context) (*defaults.Aliases, error) {
	aliases := defaults.NewAliases()
	currencyId, err := this.param.Lookup.CurrencyIdByCode(ctx, defaultCurrencyCode)
	if err != nil {
		return nil, err
	}
	aliases.Remember("currency_"+strings.ToLower(defaultCurrencyCode), currencyId)

	methods, err := this.param.PaymentMethods.ListPaymentMethods(ctx, itExt.ListPaymentMethodsQuery{})
	if err != nil {
		return nil, err
	}
	if methods.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(methods.ClientErrors.ToError(), "list payment methods")
	}
	for _, method := range methods.Data {
		// The defaults name methods in alias case (@payment_method_momo); the codes are uppercase.
		aliases.Remember(paymentMethodAlias+strings.ToLower(method.Code), model.Id(method.Id))
	}
	return aliases, nil
}

// withKnownPaymentMethods drops a link to a payment method this tenant does not have: payment
// methods are the tenant's, and a missing one is a choice, not a defect in the defaults.
func withKnownPaymentMethods(rows []dmodel.DynamicFields, aliases *defaults.Aliases) []dmodel.DynamicFields {
	kept := make([]dmodel.DynamicFields, 0, len(rows))
	for _, row := range rows {
		ref, _ := row[models.SalesChannelPaymentRelFieldPaymentMethodId].(string)
		alias := strings.TrimPrefix(ref, defaults.RefPrefix)
		if strings.HasPrefix(alias, paymentMethodAlias) {
			if _, found := aliases.Id(alias); !found {
				continue
			}
		}
		kept = append(kept, row)
	}
	return kept
}
