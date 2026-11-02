package stock

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// FindOperationTypeQuery resolves an operation type of the organization by id. A consumer keeps the
// ids it moves goods under in its own org settings, seeded when the organization is created.
type FindOperationTypeQuery struct {
	OrgId string
	Id    string
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

// StockOperationTypeReadService resolves an operation type by id. Read-only, and separate from
// StockTransferMovementService: resolving a type grants no power to move anything.
type StockOperationTypeReadService interface {
	// FindOperationTypeById returns the type, or HasData false when the organization has none
	// with that id. Absence is a normal answer, not an error.
	FindOperationTypeById(
		ctx corectx.Context, query FindOperationTypeQuery,
	) (*FindOperationTypeResult, error)
}
