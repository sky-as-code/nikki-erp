package uom

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type ConvertQuantityResult = dyn.OpResult[ConvertQuantityResultData]
type ToReferenceResult = dyn.OpResult[ToReferenceResultData]
type GetUomResult = dyn.OpResult[GetUomResultData]

// UomLookupService answers questions about a single UoM. CRUD itself belongs to the
// dynamic resource engine; this exists because other modules must be able to validate a
// UoM reference without reaching into Essential's repositories.
type UomLookupService interface {
	GetUom(ctx corectx.Context, query GetUomQuery) (*GetUomResult, error)
}

// UomConversionDomainService implements the conversion rules of BR-UOM-ESS-013.
type UomConversionDomainService interface {
	UomLookupService

	Convert(ctx corectx.Context, query ConvertQuantityQuery) (*ConvertQuantityResult, error)
	ToReference(ctx corectx.Context, query ToReferenceQuery) (*ToReferenceResult, error)
}

// UomConversionAppService is the capability other modules consume. It is the type a
// consuming module's infra/external/index.go binds to its own local port.
type UomConversionAppService interface {
	UomLookupService

	Convert(ctx corectx.Context, query ConvertQuantityQuery) (*ConvertQuantityResult, error)
	ToReference(ctx corectx.Context, query ToReferenceQuery) (*ToReferenceResult, error)
}
