package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	modconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

// NewUomDomainService is handed the composable default by the UoM onion, plus the dispatcher
// that asks the consuming modules whether a unit is still in use.
func NewUomDomainService(
	base composable.CrudDomainService, usageDispatcher *usagecheck.Dispatcher,
) itUom.UomDomainService {
	return &UomDomainServiceImpl{CrudDomainService: base, usageDispatcher: usageDispatcher}
}

// UomDomainServiceImpl attaches the UoM business invariants to the built-in create, update and
// delete. They are attached rather than reimplemented because the CRUD processing itself is
// entirely the default's; only the validation the schema cannot express belongs here.
type UomDomainServiceImpl struct {
	composable.CrudDomainService

	usageDispatcher *usagecheck.Dispatcher
}

func (this *UomDomainServiceImpl) Create(
	ctx corectx.Context, cmd itUom.CreateUomCommand, options ...composable.CreateOptions,
) (*itUom.CreateUomResult, error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, validateUomCreate(this.Repository()))
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *UomDomainServiceImpl) Update(
	ctx corectx.Context, cmd itUom.UpdateUomCommand, options ...composable.UpdateOptions,
) (*itUom.UpdateUomResult, error) {
	opts := safe.GetOptional(options, composable.UpdateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra,
		validateUomUpdate(this.Repository(), this.usageDispatcher))
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

// Delete refuses to remove a unit any consuming module still uses.
//
// The foreign keys on uom_id are ON DELETE SET NULL, so the database would accept this delete
// and quietly blank the unit from every row naming it - turning a recorded quantity into a bare
// number. That is what this check exists to prevent: the constraint keeps the reference from
// dangling, and this keeps it from being erased. A unit in use is archived, not deleted.
func (this *UomDomainServiceImpl) Delete(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.DeleteOptions,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	vErrs := ft.NewClientErrors()
	if err := assertUomDeletable(ctx, this.usageDispatcher, params, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	return this.CrudDomainService.Delete(ctx, params, options...)
}

// assertUomDeletable asks every module holding a uom_id whether this unit is still in use.
func assertUomDeletable(
	ctx corectx.Context, dispatcher *usagecheck.Dispatcher,
	params dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	uomId := readUomCatIdParam(params, models.UomFieldId)
	if uomId == "" {
		return nil
	}

	return usagecheck.AssertDeletable(ctx, dispatcher, modconstants.EssentialModuleName,
		models.UomFieldId, usagecheck.ResourceRef{
			ResourceName: usagecheck.ResourceUom,
			Identifier: usagecheck.NewIdentifier(
				uomId, readUomCatIdParam(params, basemodel.FieldOrgId)),
		}, vErrs)
}
