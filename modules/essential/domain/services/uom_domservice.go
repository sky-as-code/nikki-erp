package services

import (
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

// NewUomDomainService is handed the composable default by the UoM onion.
func NewUomDomainService(base composable.CrudDomainService) itUom.UomDomainService {
	return &UomDomainServiceImpl{CrudDomainService: base}
}

// UomDomainServiceImpl attaches the UoM business invariants to the built-in create and update.
// They are attached rather than reimplemented because the CRUD processing itself is entirely
// the default's; only the validation the schema cannot express belongs here.
type UomDomainServiceImpl struct {
	composable.CrudDomainService
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
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, validateUomUpdate(this.Repository()))
	return this.CrudDomainService.Update(ctx, cmd, opts)
}
