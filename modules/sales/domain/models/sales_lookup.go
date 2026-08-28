package models

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The reads the resource engine's built-in CRUD cannot express, kept beside the models they query
// so that the field names and the queries using them stay together.
//
// Two of them front constraints the database also enforces, as strict partial unique indexes. The
// lookups let the domain service reject a duplicate with a useful message before writing; the index
// is the backstop for the race between check and insert, not the first line of defence.

// MaxSalesPointsPerChannel bounds how many sales points one channel is read with.
//
// A channel is a classification, not a container that grows without limit: even a large vending
// estate is thousands of kiosks, not millions. The bound exists so that archiving a channel cannot
// turn one check into an unbounded read inside a transaction holding row locks.
const MaxSalesPointsPerChannel = 10000

// SalesSearcher is the slice of the repository these lookups need. Taking the narrow interface
// rather than the engine lets a unit test supply a stub without building a registry.
type SalesSearcher interface {
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error)
}

// searchAll runs a graph search and unwraps the usual three-way result into a plain slice.
func searchAll(
	ctx corectx.Context, repo SalesSearcher, graph *dmodel.SearchGraph, limit int, what string,
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

// MaxDefaultPricelistsPerOrg bounds the read that finds an organization's current default.
//
// There must be at most one (BR-PRICE-SAL-003), so this is deliberately small: it exists to tell
// "one" from "more than one" and to cap the damage if the invariant has already been broken by a
// race, not to page through a real result set.
const MaxDefaultPricelistsPerOrg = 10

// FindDefaultPricelists returns the organization's non-archived default pricelists.
//
// Plural in the name and in the return type although the answer should always be zero or one. A
// singular helper would have to decide what to do when it found two, and the only honest answers
// are to hide one or to fail — both worse than handing the caller the truth and letting the
// operation that cares decide. SetDefault uses it to demote whatever it finds, which is exactly
// the shape that repairs a broken invariant rather than tripping over it.
//
// Archived lists are excluded because an archived default is not a default: it cannot be used for
// new business, so it does not conflict with a live one (PRICE-INV-023).
func FindDefaultPricelists(
	ctx corectx.Context, repo SalesSearcher, orgId string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(SalesPricelistFieldOrgId, dmodel.Equals, orgId),
		*dmodel.NewSearchNode().NewCondition(SalesPricelistFieldIsDefault, dmodel.Equals, true),
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldIsArchived, dmodel.Equals, false),
	)
	return searchAll(ctx, repo, graph, MaxDefaultPricelistsPerOrg, "FindDefaultPricelists")
}

// CountPricelistItems reports how many rules a pricelist holds, capped.
//
// Used by the currency guard: a list that already prices something may not have its currency
// changed, because every price on it would be reinterpreted in the new one — the failure
// BR-PRICE-CUR-004 names. The caller needs only "any" or "none", so the limit is 1.
func CountPricelistItems(
	ctx corectx.Context, repo SalesSearcher, pricelistId string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			SalesPricelistItemFieldSalesPricelistId, dmodel.Equals, pricelistId),
	)
	return searchAll(ctx, repo, graph, 1, "CountPricelistItems")
}

// FindPricelistBaseOf returns a FORMULA rule of this list that derives from another list, if any.
//
// One row is enough: the caller is walking the derivation graph looking for reachability, and a
// cycle through any one base is a cycle. Which rule answers does not matter, only that some rule
// points somewhere.
func FindPricelistBaseOf(
	ctx corectx.Context, repo SalesSearcher, pricelistId string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			SalesPricelistItemFieldSalesPricelistId, dmodel.Equals, pricelistId),
		*dmodel.NewSearchNode().NewCondition(
			SalesPricelistItemFieldBasePriceSource, dmodel.Equals, PricelistBaseSourceOtherPricelist),
	)
	return searchAll(ctx, repo, graph, 1, "FindPricelistBaseOf")
}

// FindPricelistsDerivingFrom returns the lists whose FORMULA rules read the given list as their
// base, which is what a cycle check walks.
func FindPricelistsDerivingFrom(
	ctx corectx.Context, repo SalesSearcher, basePricelistId string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			SalesPricelistItemFieldBasePricelistId, dmodel.Equals, basePricelistId),
	)
	return searchAll(ctx, repo, graph, MaxSalesPointsPerChannel, "FindPricelistsDerivingFrom")
}

// FindSalesChannelByCode resolves a channel from the stable integration code an external module
// names it by (CR §22).
//
// The limit is 2 rather than 1 so that a caller can tell "exactly one" from "more than one". A
// unique constraint makes the second impossible, but a read that silently took the first row would
// hide the day that stopped being true.
func FindSalesChannelByCode(
	ctx corectx.Context, repo SalesSearcher, code string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(SalesChannelFieldCode, dmodel.Equals, code),
	)
	return searchAll(ctx, repo, graph, 2, "FindSalesChannelByCode")
}

// FindSalesPointByExternalReferenceId resolves the sales point a module registered for one of its
// own records, within one channel.
//
// This is the idempotency mechanism of CreateSalesPoint (CR §48): a vending module retrying after a
// timeout finds the point it already created instead of making a second one. The pair is unique by
// a strict partial index, so 2 is again the limit that distinguishes one from many.
//
// It filters on the id alone, not the type. Within one channel an id already identifies a point,
// and adding the type to the predicate would let a caller that passed the wrong type create a
// duplicate rather than fail — the opposite of what an idempotency lookup is for.
func FindSalesPointByExternalReferenceId(
	ctx corectx.Context, repo SalesSearcher, channelId string, externalReferenceId string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(SalesPointFieldSalesChannelId, dmodel.Equals, channelId),
		*dmodel.NewSearchNode().NewCondition(
			SalesPointFieldExternalReferenceId, dmodel.Equals, externalReferenceId),
	)
	return searchAll(ctx, repo, graph, 2, "FindSalesPointByExternalReferenceId")
}

// FindSalesPointByCode resolves a sales point by its display code within one channel, so the domain
// service can reject a duplicate before writing (D-21).
func FindSalesPointByCode(
	ctx corectx.Context, repo SalesSearcher, channelId string, code string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(SalesPointFieldSalesChannelId, dmodel.Equals, channelId),
		*dmodel.NewSearchNode().NewCondition(SalesPointFieldCode, dmodel.Equals, code),
	)
	return searchAll(ctx, repo, graph, 2, "FindSalesPointByCode")
}

// FindActiveSalesPointsOfChannel returns the points that would block archiving a channel (CR §10).
//
// "Active" here means not archived, whatever the business status: a suspended point still holds
// history and can still take returns, so archiving its channel out from under it is just as wrong
// as archiving one that is selling.
func FindActiveSalesPointsOfChannel(
	ctx corectx.Context, repo SalesSearcher, channelId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(SalesPointFieldSalesChannelId, dmodel.Equals, channelId),
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldIsArchived, dmodel.Equals, false),
	)
	return searchAll(ctx, repo, graph, limit, "FindActiveSalesPointsOfChannel")
}

// MaxPaymentMethodsPerChannel bounds how many payment mappings one channel is read with.
//
// Payment methods are configuration a human maintains, so a channel accepting more than a handful
// is already unusual. The bound exists for the same reason as MaxSalesPointsPerChannel: a listing
// must not become an unbounded read because a row was written in a loop.
const MaxPaymentMethodsPerChannel = 500

// FindPaymentMethodsOfChannel returns the mappings that say which payment methods a channel
// accepts (CR 29).
//
// The result is the local half of the merged view: it is joined against the paymentinvoice listing
// by the app service, never by the frontend. A mapping present here whose method is absent there is
// stale, not deleted (CR 34).
func FindPaymentMethodsOfChannel(
	ctx corectx.Context, repo SalesSearcher, channelId string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			SalesChannelPaymentRelFieldSalesChannelId, dmodel.Equals, channelId),
	)
	return searchAll(ctx, repo, graph, MaxPaymentMethodsPerChannel, "FindPaymentMethodsOfChannel")
}

// FindChannelPaymentMapping resolves the single mapping of one channel to one method, which is
// what makes enabling idempotent and disabling safe to repeat.
//
// The limit is 2 for the same reason as the other lookups: a composite unique makes a second row
// impossible, and a read that silently took the first would hide the day that stopped being true.
func FindChannelPaymentMapping(
	ctx corectx.Context, repo SalesSearcher, channelId string, paymentMethodId string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			SalesChannelPaymentRelFieldSalesChannelId, dmodel.Equals, channelId),
		*dmodel.NewSearchNode().NewCondition(
			SalesChannelPaymentRelFieldPaymentMethodId, dmodel.Equals, paymentMethodId),
	)
	return searchAll(ctx, repo, graph, 2, "FindChannelPaymentMapping")
}
