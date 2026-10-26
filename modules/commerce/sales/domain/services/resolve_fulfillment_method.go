package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Which fulfillment policy an order is sold under, and where its goods will come from.
//
// The answer is settled ONCE, at the moment of sale, and everything downstream reads the snapshot
// the fulfillment took rather than asking again. That is the point of resolving here: a customer
// buys under the policy that was in force when they paid, and an operator editing the catalogue
// afterwards must not be able to change what a live order promised them.

const (
	ReasonMethodUnresolved    = "sales_fulfillment.method_unresolved"
	ReasonMethodRequiresAuth  = "sales_fulfillment.method_requires_authenticated_customer"
	ReasonTargetRequired      = "sales_fulfillment.target_required"
	ReasonTargetNotFound      = "sales_fulfillment.target_not_found"
	ReasonTargetNotFulfilling = "sales_fulfillment.target_not_fulfillment_enabled"
	ReasonTargetHasNoLocation = "sales_fulfillment.target_has_no_inventory_location"
)

// FulfillmentMethodRequest is what a caller knows when it asks. Both ids are optional: the whole
// job of resolution is to fill in what the caller did not name.
type FulfillmentMethodRequest struct {
	// FulfillmentMethodId is the method the client asked for outright, and takes precedence over
	// every default. Empty when the client did not choose.
	FulfillmentMethodId string

	// TargetOutletId is the sales point the client wants the goods handed over at. Empty unless
	// the client chose, which only a customer_selected_outlet method requires them to do.
	TargetOutletId string

	// The order's own facts, read from the order rather than accepted from the client.
	SalesChannelId       string
	SalesPointId         string
	OrgId                string
	CustomerIdentityMode models.CustomerIdentityMode
}

// ResolvedFulfillmentMethod is the settled policy plus the target it will be executed at. Every
// field on it is copied onto the fulfillment as a snapshot; nothing downstream re-reads the method.
type ResolvedFulfillmentMethod struct {
	MethodId       string
	MethodCode     string
	Type           models.FulfillmentType
	TargetOutletId string

	// TargetLocationId is the inventory location behind the target point. Resolved here so the
	// reservation names a place rather than letting Inventory infer one from an operation type's
	// defaults, which would hold stock somewhere the customer was never promised.
	TargetLocationId string

	MaxAttempts             *int32
	FailureAction           models.FulfillmentFailureAction
	AllowTargetChange       bool
	AllowPartialFulfillment bool
	ReservationTtlMinutes   *int32
}

// ResolveFulfillmentMethod settles the policy and the target for one order.
//
// The method is looked for in the order the CR fixes: what the client asked for, then the sales
// point's default, then the channel's. Precedence is not preference — it is specificity. A client
// naming a method has made a choice about this sale; a point default describes how that one machine
// sells; a channel default is the fallback for everything else.
//
// A resolution that finds nothing is a refusal rather than a silent skip: an order with no policy
// would confirm, take money, and then have no rule saying what happens when the goods do not come
// out, which is precisely the situation this whole feature exists to prevent.
func ResolveFulfillmentMethod(
	ctx corectx.Context,
	request FulfillmentMethodRequest,
	methods *SalesFulfillmentMethodDomainServiceImpl,
) (*ResolvedFulfillmentMethod, *dyn.OpResult[dyn.MutateResultData], error) {
	methodId, refusal, err := resolveMethodId(ctx, request)
	if err != nil || refusal != nil {
		return nil, refusal, err
	}

	// The shared gate first: exists, not archived, an implemented type, allowed on this channel.
	// Running it here rather than re-checking those four conditions means a rule added to it later
	// applies to order creation without this file being touched.
	if methods != nil {
		assignable, err := methods.AssertAssignable(ctx, methodId, request.SalesChannelId)
		if err != nil {
			return nil, nil, err
		}
		if assignable != nil && assignable.ClientErrors.Count() > 0 {
			return nil, assignable, nil
		}
	}

	method, err := loadRecord(ctx, models.SalesFulfillmentMethodSchemaName,
		models.SalesFulfillmentMethodFieldId, methodId)
	if err != nil {
		return nil, nil, err
	}
	if method == nil {
		return nil, violationResult(models.SalesFulfillmentMethodSchemaName,
			ReasonMethodNotFound, "no fulfillment method with id "+methodId), nil
	}
	if orgId := stringOf(method, basemodel.FieldOrgId); !sameOrg(orgId, request.OrgId) {
		// Answered as "no such method" rather than "not yours": telling a caller that an id they
		// guessed exists in another organization is the leak this refusal exists to prevent.
		return nil, violationResult(models.SalesFulfillmentMethodSchemaName,
			ReasonMethodNotFound, "no fulfillment method with id "+methodId), nil
	}

	// An identity requirement is checked against the order's own snapshot, never against the live
	// request: the order records what was verified when it was raised, and a later call carrying a
	// different principal must not requalify a sale that was made anonymously.
	if boolOf(method, models.SalesFulfillmentMethodFieldRequiresAuthenticatedCustomer) &&
		request.CustomerIdentityMode != models.CustomerIdentityModeAuthenticated {
		return nil, violationResult(models.SalesFulfillmentMethodSchemaName,
			ReasonMethodRequiresAuth,
			"this fulfillment method may only be used by an identified customer, and this order "+
				"was raised anonymously"), nil
	}

	targetId, refusal, err := resolveTargetOutletId(request, method)
	if err != nil || refusal != nil {
		return nil, refusal, err
	}

	locationId, refusal, err := assertTargetUsable(ctx, targetId, request.OrgId)
	if err != nil || refusal != nil {
		return nil, refusal, err
	}

	return &ResolvedFulfillmentMethod{
		MethodId:                methodId,
		MethodCode:              stringOf(method, models.SalesFulfillmentMethodFieldCode),
		Type:                    models.FulfillmentType(stringOf(method, models.SalesFulfillmentMethodFieldFulfillmentType)),
		TargetOutletId:          targetId,
		TargetLocationId:        locationId,
		MaxAttempts:             optionalInt32Of(method, models.SalesFulfillmentMethodFieldMaxAttempts),
		FailureAction:           models.FulfillmentFailureAction(stringOf(method, models.SalesFulfillmentMethodFieldFailureAction)),
		AllowTargetChange:       boolOf(method, models.SalesFulfillmentMethodFieldAllowTargetChange),
		AllowPartialFulfillment: boolOf(method, models.SalesFulfillmentMethodFieldAllowPartialFulfillment),
		ReservationTtlMinutes:   optionalInt32Of(method, models.SalesFulfillmentMethodFieldReservationTtlMinutes),
	}, nil, nil
}

// resolveMethodId walks explicit, then point default, then channel default.
func resolveMethodId(
	ctx corectx.Context, request FulfillmentMethodRequest,
) (string, *dyn.OpResult[dyn.MutateResultData], error) {
	if request.FulfillmentMethodId != "" {
		return request.FulfillmentMethodId, nil, nil
	}

	if request.SalesPointId != "" {
		point, err := loadRecord(ctx, models.SalesPointSchemaName,
			models.SalesPointFieldId, request.SalesPointId)
		if err != nil {
			return "", nil, err
		}
		if methodId := stringOf(point, models.SalesPointFieldDefaultFulfillmentMethodId); methodId != "" {
			return methodId, nil, nil
		}
	}

	if request.SalesChannelId != "" {
		channel, err := loadRecord(ctx, models.SalesChannelSchemaName,
			models.SalesChannelFieldId, request.SalesChannelId)
		if err != nil {
			return "", nil, err
		}
		if methodId := stringOf(channel, models.SalesChannelFieldDefaultFulfillmentMethodId); methodId != "" {
			return methodId, nil, nil
		}
	}

	return "", violationResult(models.SalesFulfillmentMethodSchemaName,
		ReasonMethodUnresolved,
		"no fulfillment method was named on the order, and neither the sales point nor the "+
			"sales channel declares a default"), nil
}

// resolveTargetOutletId applies the method's selection strategy.
//
// A customer_selected_outlet method that was given no target is refused rather than defaulted to
// where the order was raised: guessing which kiosk a customer meant to collect from would send
// their goods to another town, and the refusal is recoverable while a wrong reservation is not.
func resolveTargetOutletId(
	request FulfillmentMethodRequest, method dmodel.DynamicFields,
) (string, *dyn.OpResult[dyn.MutateResultData], error) {
	if request.TargetOutletId != "" {
		return request.TargetOutletId, nil, nil
	}

	selection := models.InitialTargetSelection(
		stringOf(method, models.SalesFulfillmentMethodFieldInitialTargetSelection))
	switch selection {
	case models.InitialTargetSelectionCurrentSalesOutlet:
		if request.SalesPointId == "" {
			return "", violationResult(models.SalesOrderFulfillmentSchemaName,
				ReasonTargetRequired,
				"this fulfillment method delivers from the sales point the order was raised at, "+
					"and the order names none"), nil
		}
		return request.SalesPointId, nil, nil

	case models.InitialTargetSelectionCustomerSelectedOutlet:
		return "", violationResult(models.SalesOrderFulfillmentSchemaName,
			ReasonTargetRequired,
			"this fulfillment method requires the customer to choose where to collect; name a "+
				"target sales point before confirming the order"), nil
	}

	return "", violationResult(models.SalesOrderFulfillmentSchemaName,
		ReasonTargetRequired,
		"the fulfillment method declares no target selection strategy"), nil
}

// assertTargetUsable checks the point can actually execute a fulfillment, and returns the inventory
// location behind it.
//
// The location is required rather than optional because a fulfillment-enabled point with none would
// confirm an order and then have nowhere to hold its stock — a promise with nothing behind it. The
// schema cannot express the pairing (it is a business rule, not a column constraint), so it is
// enforced at every point of use instead.
func assertTargetUsable(
	ctx corectx.Context, targetId string, orgId string,
) (string, *dyn.OpResult[dyn.MutateResultData], error) {
	if targetId == "" {
		return "", violationResult(models.SalesOrderFulfillmentSchemaName,
			ReasonTargetRequired, "no fulfillment target was resolved"), nil
	}

	point, err := loadRecord(ctx, models.SalesPointSchemaName, models.SalesPointFieldId, targetId)
	if err != nil {
		return "", nil, err
	}
	// Cross-org and non-existent answer identically. A caller must not be able to discover that an
	// id is real by the wording of its refusal.
	if point == nil || !sameOrg(stringOf(point, basemodel.FieldOrgId), orgId) {
		return "", violationResult(models.SalesPointSchemaName,
			ReasonTargetNotFound, "no sales point with id "+targetId), nil
	}
	if !boolOf(point, models.SalesPointFieldFulfillmentEnabled) {
		return "", violationResult(models.SalesPointSchemaName,
			ReasonTargetNotFulfilling,
			"sales point "+targetId+" is not enabled for fulfillment"), nil
	}

	locationId := stringOf(point, models.SalesPointFieldInventoryLocationId)
	if locationId == "" {
		return "", violationResult(models.SalesPointSchemaName,
			ReasonTargetHasNoLocation,
			"sales point "+targetId+" is enabled for fulfillment but names no inventory "+
				"location, so there is nowhere to hold its stock"), nil
	}
	return locationId, nil, nil
}

// sameOrg treats an unknown org on either side as a match, because org scoping is enforced by the
// engine pipeline before this code runs; this is the second line of defence against an id that
// arrived through a path that did not scope, not the first.
func sameOrg(recordOrgId, requestOrgId string) bool {
	return recordOrgId == "" || requestOrgId == "" || recordOrgId == requestOrgId
}

func optionalInt32Of(record dmodel.DynamicFields, field string) *int32 {
	if record == nil {
		return nil
	}
	value, ok := record[field]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case int32:
		return &typed
	case *int32:
		return typed
	case int:
		narrowed := int32(typed)
		return &narrowed
	case int64:
		narrowed := int32(typed)
		return &narrowed
	case float64:
		narrowed := int32(typed)
		return &narrowed
	}
	return nil
}
