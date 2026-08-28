package models

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// MaxOrderLines bounds how many lines one purchase order is read with.
//
// An order is a document a buyer fills in, not a bulk import, so a thousand lines is already far
// past anything a purchasing department produces by hand. The bound exists so that a corrupted or
// malicious order cannot turn one recompute into an unbounded read inside a transaction that is
// holding row locks — the same reasoning as inventory's MaxTransferMoves.
const MaxOrderLines = 1000

// MaxAgreementLines bounds how many lines one agreement is read with, for the same reason.
const MaxAgreementLines = 1000

// PurchaseSearcher is the slice of the repository these lookups need. Taking the narrow interface
// rather than the engine lets a unit test supply a stub without building a registry.
type PurchaseSearcher interface {
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error)
}

// searchAll runs a graph search and unwraps the usual three-way result into a plain slice.
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

// FindOrderLines returns the lines of one purchase order.
//
// This is the read the totals recomputation is built on: the lines are what a reader can verify by
// adding up, so they are the authority for the header's three totals rather than the other way
// around (PUR-R4, D8).
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
//
// It searches rather than fetching by key because org_id is a unique index and not the primary
// key. The limit is 2 so that a caller can tell "exactly one" from "more than one" — the unique
// index makes the second impossible, but a read that silently takes the first row would hide the
// day it stops being true.
func FindConfigurationForOrg(
	ctx corectx.Context, repo PurchaseSearcher, orgId string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(ConfigurationFieldOrgId, dmodel.Equals, orgId),
	)
	return searchAll(ctx, repo, graph, 2, "FindConfigurationForOrg")
}

// MaxAgreementOrders bounds how many orders one agreement's drawdown is derived from.
//
// An agreement accumulates orders over its life, so this is larger than the per-document bounds:
// what it guards against is an unbounded read inside a transaction, not a document nobody would
// type by hand.
const MaxAgreementOrders = 5000

// FindOpenOrdersForAgreement returns orders drawn against an agreement that are neither confirmed
// nor cancelled — the ones that would be stranded if the agreement closed (BR §42).
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

// FindConfirmedOrdersForAgreement returns the orders that count towards an agreement's drawdown.
//
// Only confirmed ones: a draft RFQ is a question rather than a commitment, and counting it would
// show an agreement as drawn down by orders that may never be placed (BR §41).
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

// MaxSourcingGroupOrders bounds how many alternatives one requirement is compared across.
//
// Asking more than a handful of vendors for the same thing is already unusual; this is a guard
// against a runaway read rather than a business limit.
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

// MaxVendorPriceCandidates bounds how many vendor price rows one line is priced from.
//
// The read is already narrowed to one vendor and one product template, so a realistic figure is a
// handful — a few quantity breaks, perhaps one row per variant. The bound is a guard against a
// vendor whose price list has been imported badly, not a business limit: resolution is a pure
// function over whatever it is handed, and handing it ten thousand rows inside a write transaction
// would be the problem long before choosing between them was.
const MaxVendorPriceCandidates = 200

// FindVendorPriceCandidates returns every non-archived price this vendor offers for this product
// template, in this organization.
//
// It deliberately does NOT filter by variant, by quantity or by date. Those three are resolution's
// job (section 27) and the reasons differ:
//
//   - VARIANT, because a template-wide row prices a variant that has no row of its own, so
//     filtering to the variant would discard the row that ought to win.
//   - QUANTITY, because the winner is the highest break the request REACHES, which cannot be
//     expressed as a filter without also ordering — and the ordering is what the resolver exists to
//     decide.
//   - DATE, because the validity verdict belongs to the caller's pricing date, and the resolver
//     takes it as a boolean precisely so it can stay free of a clock.
//
// Archived rows ARE excluded, and that is the one filter that belongs here: an archived quote is
// withdrawn and must never price something new. It stays readable — a confirmed order that resolved
// through it still names it (PRICE-INV-024) — but reading it back is a different operation from
// pricing with it.
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
