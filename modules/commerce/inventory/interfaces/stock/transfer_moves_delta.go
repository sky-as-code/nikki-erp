package stock

import (
	"encoding/json"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
)

// TransferFieldMoves carries the moves delta on a transfer create or update. It is not a transfer
// column: the domain service lifts it before the schema sees the body.
const TransferFieldMoves = "moves"

// TransferMovesDeltaMaxItems caps each list of the delta: every item is a read and a write inside
// the transfer's transaction.
const TransferMovesDeltaMaxItems = 200

// TransferMoveAdd is one move to create. Locations, status and sequence come from the transfer.
// A nil UomId means the variant's base unit.
type TransferMoveAdd struct {
	ProductVariantId model.Id         `json:"product_variant_id"`
	UomId            *model.Id        `json:"uom_id"`
	DemandQuantity   *decimal.Decimal `json:"demand_quantity"`
}

// TransferMoveUpdate changes one move the transfer already holds. A changed variant replaces the
// move, because a move's variant cannot be updated in place.
type TransferMoveUpdate struct {
	Id               model.Id         `json:"id"`
	Etag             model.Etag       `json:"etag"`
	ProductVariantId *model.Id        `json:"product_variant_id"`
	UomId            *model.Id        `json:"uom_id"`
	DemandQuantity   *decimal.Decimal `json:"demand_quantity"`
}

// TransferMovesDelta rides a transfer create or update as `moves: {add, update, remove}`, the same
// shape as the IAM role's `entitlements`.
type TransferMovesDelta struct {
	Add    []TransferMoveAdd    `json:"add"`
	Update []TransferMoveUpdate `json:"update"`
	Remove []model.Id           `json:"remove"`
}

func (this TransferMovesDelta) IsEmpty() bool {
	return len(this.Add) == 0 && len(this.Update) == 0 && len(this.Remove) == 0
}

// TransferMovesDeltaParts reports which lists a delta carries, so each can be authorized as the
// matching stock-move permission.
type TransferMovesDeltaParts struct {
	Add    bool
	Update bool
	Remove bool
}

// ReadTransferMovesDeltaParts reads the delta without changing the command. A malformed delta
// counts as every part: it must still be authorized as one.
func ReadTransferMovesDeltaParts(fields dmodel.DynamicFields) TransferMovesDeltaParts {
	raw, found := fields[TransferFieldMoves]
	if !found || raw == nil {
		return TransferMovesDeltaParts{}
	}
	delta, err := decodeTransferMovesDelta(raw)
	if err != nil {
		return TransferMovesDeltaParts{Add: true, Update: true, Remove: true}
	}
	return TransferMovesDeltaParts{
		Add:    len(delta.Add) > 0,
		Update: len(delta.Update) > 0,
		Remove: len(delta.Remove) > 0,
	}
}

// LiftMovesDelta removes `moves` from a transfer command, so the transfer schema never sees it,
// and returns the decoded delta. malformed reports a value that is not a delta.
func LiftMovesDelta(fields dmodel.DynamicFields) (delta *TransferMovesDelta, malformed bool) {
	raw, found := fields[TransferFieldMoves]
	if !found {
		return nil, false
	}
	delete(fields, TransferFieldMoves)
	if raw == nil {
		return nil, false
	}
	delta, err := decodeTransferMovesDelta(raw)
	if err != nil {
		return nil, true
	}
	return delta, false
}

func decodeTransferMovesDelta(raw any) (*TransferMovesDelta, error) {
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	delta := &TransferMovesDelta{}
	if err := json.Unmarshal(encoded, delta); err != nil {
		return nil, err
	}
	return delta, nil
}
