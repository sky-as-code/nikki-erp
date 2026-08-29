package models

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The reads the resource engine's built-in CRUD cannot express. Two of them front constraints the
// database also enforces as strict partial unique indexes: the lookup gives a useful message before
// writing, the index is the backstop for the race between check and insert.

// MaxSalesPointsPerChannel bounds how many sales points one channel is read with, so archiving a
// channel cannot become an unbounded read inside a transaction holding row locks.
const MaxSalesPointsPerChannel = 10000

// SalesSearcher is the narrow slice of the repository these lookups need, so a unit test can supply
// a stub without building a registry.
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

// MaxDefaultPricelistsPerOrg bounds the read that finds an organization's current default. There
// must be at most one, so this only tells "one" from "more than one".
const MaxDefaultPricelistsPerOrg = 10

// FindDefaultPricelists returns the organization's non-archived default pricelists. Plural although
// the answer should be zero or one: SetDefault demotes whatever it finds, which repairs a broken
// invariant rather than tripping over it. Archived lists are excluded because an archived default
// cannot be used for new business and so does not conflict with a live one.
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

// CountPricelistItems reports whether a pricelist holds any rules; the limit is 1 because the
// currency guard needs only "any" or "none". A list that already prices something may not have its
// currency changed, since every price on it would be reinterpreted in the new one.
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

// FindPricelistBaseOf returns a formula rule of this list that derives from another list, if any.
// One row is enough: a cycle through any one base is a cycle.
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
// names it by. The limit is 2 so a caller can tell "exactly one" from "more than one": a unique
// constraint makes the second impossible, but taking the first row would hide the day it was not.
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
// own records, within one channel. This is CreateSalesPoint's idempotency mechanism: a caller
// retrying after a timeout finds the point it already created. It filters on the id alone, not the
// type — adding the type would let a caller passing the wrong one create a duplicate rather than fail.
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
// service can reject a duplicate before writing.
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

// FindActiveSalesPointsOfChannel returns the points that would block archiving a channel. "Active"
// means not archived, whatever the business status: a suspended point still holds history and takes
// returns.
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

// MaxPaymentMethodsPerChannel bounds how many payment mappings one channel is read with, so a
// listing cannot become an unbounded read because rows were written in a loop.
const MaxPaymentMethodsPerChannel = 500

// FindPaymentMethodsOfChannel returns the mappings that say which payment methods a channel accepts.
// This is the local half of a merged view, joined against the paymentinvoice listing by the app
// service and never by the frontend. A mapping whose method is absent there is stale, not deleted.
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

// FindChannelPaymentMapping resolves the single mapping of one channel to one method, which makes
// enabling idempotent and disabling safe to repeat. The limit is 2 so a second row surfaces rather
// than being silently ignored.
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
