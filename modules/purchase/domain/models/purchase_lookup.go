package models

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// MaxOrderLines bounds how many lines one purchase order is read with, so a corrupted or malicious
// order cannot turn one recompute into an unbounded read inside a lock-holding transaction.
const MaxOrderLines = 1000

// MaxAgreementLines bounds how many lines one agreement is read with, for the same reason.
const MaxAgreementLines = 1000

// PurchaseSearcher is the narrow repository slice these lookups need, so a unit test can supply a
// stub without building a registry.
type PurchaseSearcher interface {
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error)
}

// searchAll unwraps the three-way search result into a plain slice.
func searchAll(
	ctx corectx.Context, repo PurchaseSearcher, graph *dmodel.SearchGraph, limit int, what string,
) ([]dmodel.DynamicFields, error) {
	found, err := repo.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  limit,
	})
	if err != nil {
		return nil, errors.Wrap(err, what)
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), what)
	}
	if !found.HasData {
		return nil, nil
	}
	return found.Data.Items, nil
}

// FindOrderLines returns the lines of one purchase order. Lines are the authority for the header's
// totals, not the other way around.
func FindOrderLines(
	ctx corectx.Context, repo PurchaseSearcher, orderId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(PurchaseOrderLineFieldPurchaseOrderId, dmodel.Equals, orderId),
	)
	return searchAll(ctx, repo, graph, limit, "FindOrderLines")
}

// FindAgreementLines returns the lines of one purchase agreement.
func FindAgreementLines(
	ctx corectx.Context, repo PurchaseSearcher, agreementId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(AgreementLineFieldPurchaseAgreementId, dmodel.Equals, agreementId),
	)
	return searchAll(ctx, repo, graph, limit, "FindAgreementLines")
}

// FindConfigurationForOrg returns the purchase configuration of one organization, if it has one.
// It searches because org_id is a unique index and not the primary key; the limit is 2 so a caller
// can tell "exactly one" from "more than one" instead of silently taking the first row.
func FindConfigurationForOrg(
	ctx corectx.Context, repo PurchaseSearcher, orgId string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(ConfigurationFieldOrgId, dmodel.Equals, orgId),
	)
	return searchAll(ctx, repo, graph, 2, "FindConfigurationForOrg")
}

// MaxAgreementOrders bounds how many orders one agreement's drawdown is derived from. It is larger
// than the per-document bounds because it guards an unbounded read inside a transaction.
const MaxAgreementOrders = 5000

// FindOpenOrdersForAgreement returns orders drawn against an agreement that are neither confirmed
// nor cancelled — the ones that would be stranded if the agreement closed.
func FindOpenOrdersForAgreement(
	ctx corectx.Context, repo PurchaseSearcher, agreementId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			PurchaseOrderFieldAgreementId, dmodel.Equals, agreementId),
		*dmodel.NewSearchNode().NewCondition(
			PurchaseOrderFieldStatus, dmodel.In, []string{
				string(PurchaseOrderStatusRfq),
				string(PurchaseOrderStatusRfqSent),
				string(PurchaseOrderStatusToApprove),
			}),
	)
	return searchAll(ctx, repo, graph, limit, "FindOpenOrdersForAgreement")
}

// FindConfirmedOrdersForAgreement returns the orders counting towards an agreement's drawdown.
// Only confirmed ones: a draft RFQ is a question, not a commitment.
func FindConfirmedOrdersForAgreement(
	ctx corectx.Context, repo PurchaseSearcher, agreementId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			PurchaseOrderFieldAgreementId, dmodel.Equals, agreementId),
		*dmodel.NewSearchNode().NewCondition(
			PurchaseOrderFieldStatus, dmodel.Equals, string(PurchaseOrderStatusPurchaseOrder)),
	)
	return searchAll(ctx, repo, graph, limit, "FindConfirmedOrdersForAgreement")
}

// MaxSourcingGroupOrders bounds how many alternatives one requirement is compared across. It is a
// guard against a runaway read, not a business limit.
const MaxSourcingGroupOrders = 100

// FindOrdersInSourcingGroup returns every order being compared as an alternative to the others.
func FindOrdersInSourcingGroup(
	ctx corectx.Context, repo PurchaseSearcher, groupId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			PurchaseOrderFieldSourcingGroupId, dmodel.Equals, groupId),
	)
	return searchAll(ctx, repo, graph, limit, "FindOrdersInSourcingGroup")
}

// MaxVendorPriceCandidates bounds how many vendor price rows one line is priced from. It guards
// against a badly imported price list flooding a write transaction, not a business limit.
const MaxVendorPriceCandidates = 200

// FindVendorPriceCandidates returns every non-archived price this vendor offers for this product
// template, in this organization.
//
// It deliberately does not filter by variant, quantity or date — those are the resolver's job: a
// template-wide row may price a variant that has no row of its own, the winning quantity break is
// the highest one reached rather than a filter, and validity is judged against the caller's pricing
// date so the resolver stays free of a clock. Archived rows are excluded because a withdrawn quote
// must never price something new, though it stays readable for orders that already resolved to it.
func FindVendorPriceCandidates(
	ctx corectx.Context, repo PurchaseSearcher, orgId, vendorId, templateId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(VendorProductPriceFieldOrgId, dmodel.Equals, orgId),
		*dmodel.NewSearchNode().NewCondition(VendorProductPriceFieldVendorId, dmodel.Equals, vendorId),
		*dmodel.NewSearchNode().NewCondition(
			VendorProductPriceFieldProductTemplateId, dmodel.Equals, templateId),
		*dmodel.NewSearchNode().NewCondition(VendorProductPriceFieldIsArchived, dmodel.Equals, false),
	)
	return searchAll(ctx, repo, graph, limit, "FindVendorPriceCandidates")
}
