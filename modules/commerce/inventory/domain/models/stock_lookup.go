package models

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// MaxTransferMoves bounds how many moves one transfer is read with, so a corrupted or malicious
// transfer cannot turn one validate into an unbounded read inside a transaction holding row locks.
const MaxTransferMoves = 1000

// MaxMoveLines bounds how many move lines one move is read with, for the same reason as
// MaxTransferMoves.
const MaxMoveLines = 1000

// FindTransferMoves returns the moves of a transfer, in sequence order.
func FindTransferMoves(
	ctx corectx.Context, repo ProductSearcher, transferId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(StockMoveFieldTransferId, dmodel.Equals, transferId),
	)
	return searchAll(ctx, repo, graph, limit, "FindTransferMoves")
}

// FindMoveLines returns a move's lines: its reservations until validate stamps them, its executed
// detail afterwards, so both reserve and validate read through here.
func FindMoveLines(
	ctx corectx.Context, repo ProductSearcher, moveId string, limit int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(StockMoveLineFieldMoveId, dmodel.Equals, moveId),
	)
	return searchAll(ctx, repo, graph, limit, "FindMoveLines")
}

// FindTransferByIdempotencyKey returns the transfer already carrying a validate key, if any: a
// retried validate finds its earlier result here instead of moving the stock a second time. Scoped
// to an org because the partial unique on the key is per-org.
func FindTransferByIdempotencyKey(
	ctx corectx.Context, repo ProductSearcher, orgId string, key string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(StockTransferFieldOrgId, dmodel.Equals, orgId),
		*dmodel.NewSearchNode().NewCondition(StockTransferFieldIdempotencyKey, dmodel.Equals, key),
	)
	return searchAll(ctx, repo, graph, 2, "FindTransferByIdempotencyKey")
}

// FindQuantForDimension returns the quant for an exact variant/location/lot/package/owner key.
// Validate's destination balance usually does not exist yet, so the caller must create one when
// this returns nothing rather than assume an update will land.
func FindQuantForDimension(
	ctx corectx.Context, repo ProductSearcher, key QuantDimension,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(StockQuantFieldOrgId, dmodel.Equals, key.OrgId),
		*dmodel.NewSearchNode().NewCondition(StockQuantFieldProductVariantId, dmodel.Equals, key.ProductVariantId),
		*dmodel.NewSearchNode().NewCondition(StockQuantFieldLocationId, dmodel.Equals, key.LocationId),
		*dmodel.NewSearchNode().NewCondition(StockQuantFieldLotRef, dmodel.Equals, key.LotRef),
		*dmodel.NewSearchNode().NewCondition(StockQuantFieldPackageRef, dmodel.Equals, key.PackageRef),
		*dmodel.NewSearchNode().NewCondition(StockQuantFieldOwnerRef, dmodel.Equals, key.OwnerRef),
	)
	return searchAll(ctx, repo, graph, 1, "FindQuantForDimension")
}

// FindQuantsAtLocation returns every balance of a variant in a location, across all lots, packages
// and owners. Check-availability reads through here rather than the locking helper because it
// deliberately takes no lock: the figure is a snapshot another request may invalidate at once.
func FindQuantsAtLocation(
	ctx corectx.Context, repo ProductSearcher, orgId, variantId, locationId string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(StockQuantFieldOrgId, dmodel.Equals, orgId),
		*dmodel.NewSearchNode().NewCondition(StockQuantFieldProductVariantId, dmodel.Equals, variantId),
		*dmodel.NewSearchNode().NewCondition(StockQuantFieldLocationId, dmodel.Equals, locationId),
	)
	return searchAll(ctx, repo, graph, MaxMoveLines, "FindQuantsAtLocation")
}

// QuantDimension is the full unique key of a stock balance. The three refs are empty strings rather
// than nulls, matching the schema.
type QuantDimension struct {
	OrgId            string
	ProductVariantId string
	LocationId       string
	LotRef           string
	PackageRef       string
	OwnerRef         string
}
