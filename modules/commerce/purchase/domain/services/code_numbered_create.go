package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// createWithCode stamps the record's code from its series and creates it, in one transaction: the
// code is reserved before the row is written, so a refused write releases it. Whatever code the
// client sent is overwritten, as it always was.
func createWithCode(
	ctx corectx.Context, base dynamicresource.CrudDomainService, numbering *CodeNumbering,
	sequence CodeSequence, prepared dmodel.DynamicFields, options ...dynamicresource.CreateOptions,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	if numbering == nil {
		return nil, errors.Errorf("create with %s: code numbering is not installed", sequence.Field)
	}
	var result *dyn.OpResult[dmodel.DynamicFields]
	refused, err := dynamicresource.InRefusableTransaction(ctx, base.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			orgId := model.Id(stringOf(prepared, basemodel.FieldOrgId))
			code, cErrs, err := numbering.Next(tranxCtx, sequence, orgId)
			if err != nil || cErrs != nil {
				return cErrs, err
			}
			prepared[sequence.Field] = code
			result, err = base.Create(tranxCtx, prepared, options...)
			if err != nil {
				return nil, err
			}
			return &result.ClientErrors, nil
		})
	if err != nil {
		return nil, errors.Wrapf(err, "create with %s", sequence.Field)
	}
	if refused != nil {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *refused}, nil
	}
	return result, nil
}
