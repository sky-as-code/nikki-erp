package stock

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// FindOperationTypeQuery resolves an operation type by the code a consumer knows at compile time,
// so a module that always moves goods the same way never has to be configured with an id. Ids are
// per organization; a code is not, so an id would need one per organization and would silently be
// missing for one added later.
type FindOperationTypeQuery struct {
	OrgId string

	// Code is the operation type's own `code`, unique within the organization — not the
	// operation_code enum, which many types share.
	Code string
}

// OperationTypeRef is the little a mover needs: enough to raise a transfer against the type and to
// check it is pointed the way the caller expects. The policies are deliberately absent — they are
// snapshotted onto the transfer at create, so reading them here would invite a caller to
// second-guess a decision Stock has already made.
type OperationTypeRef struct {
	Id string

	// OperationCode is incoming, outgoing or internal.
	OperationCode string

	IsArchived bool
}

type FindOperationTypeResultData struct {
	OperationType OperationTypeRef
}

type FindOperationTypeResult = dyn.OpResult[FindOperationTypeResultData]

// StockOperationTypeReadService resolves an operation type by code. Read-only, and separate from
// StockTransferMovementService: resolving a type grants no power to move anything.
type StockOperationTypeReadService interface {
	// FindOperationTypeByCode returns the type, or HasData false when the organization has none
	// with that code. Absence is a normal answer, not an error.
	FindOperationTypeByCode(
		ctx corectx.Context, query FindOperationTypeQuery,
	) (*FindOperationTypeResult, error)
}
