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
// Two of these back constraints the database does not enforce. The sales point uniqueness rules are
// partial indexes written by hand in the migration rather than declared in the schema (see
// sales_point.json), so the lookups below are how the domain service checks them before writing —
// the index is the backstop, not the first line of defence.

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

// FindSalesPointByExternalReference resolves the sales point a module registered for one of its own
// records, within one channel.
//
// This is the idempotency mechanism of CreateSalesPoint (CR §48): a vending module retrying after a
// timeout finds the point it already created instead of making a second one. The pair is unique by
// a partial index in the migration, so 2 is again the limit that distinguishes one from many.
func FindSalesPointByExternalReference(
	ctx corectx.Context, repo SalesSearcher, channelId string, externalReference string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(SalesPointFieldSalesChannelId, dmodel.Equals, channelId),
		*dmodel.NewSearchNode().NewCondition(
			SalesPointFieldExternalReference, dmodel.Equals, externalReference),
	)
	return searchAll(ctx, repo, graph, 2, "FindSalesPointByExternalReference")
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
