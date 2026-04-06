package repository

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

const authorizeModuleName = "authorize"

type crudContextAsCore struct {
	crud.Context
}

func (crudContextAsCore) GetModuleName() string {
	return authorizeModuleName
}

func (crudContextAsCore) GetDomainConstraints() dmodel.DynamicFields {
	return nil
}

// ToCoreCtx adapts crud.Context for code that requires corectx.Context (e.g. dynamic baserepo).
func ToCoreCtx(ctx crud.Context) corectx.Context {
	if ctx == nil {
		return nil
	}
	if c, ok := ctx.(corectx.Context); ok {
		return c
	}
	return crudContextAsCore{Context: ctx}
}
