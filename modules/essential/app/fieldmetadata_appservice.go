package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/fieldmetadata"
)

func NewFieldMetadataApplicationServiceImpl(fieldMetadataSvc it.FieldMetadataDomainService) it.FieldMetadataAppService {
	return &FieldMetadataApplicationServiceImpl{fieldMetadataSvc: fieldMetadataSvc}
}

type FieldMetadataApplicationServiceImpl struct {
	fieldMetadataSvc it.FieldMetadataDomainService
}

func (this *FieldMetadataApplicationServiceImpl) CreateFieldMetadata(ctx corectx.Context, cmd it.CreateFieldMetadataCommand) (*it.CreateFieldMetadataResult, error) {
	return this.fieldMetadataSvc.CreateFieldMetadata(ctx, cmd)
}

func (this *FieldMetadataApplicationServiceImpl) DeleteFieldMetadata(ctx corectx.Context, cmd it.DeleteFieldMetadataCommand) (*it.DeleteFieldMetadataResult, error) {
	return this.fieldMetadataSvc.DeleteFieldMetadata(ctx, cmd)
}

func (this *FieldMetadataApplicationServiceImpl) FieldMetadataExists(ctx corectx.Context, query it.FieldMetadataExistsQuery) (*it.FieldMetadataExistsResult, error) {
	return this.fieldMetadataSvc.FieldMetadataExists(ctx, query)
}

func (this *FieldMetadataApplicationServiceImpl) GetFieldMetadata(ctx corectx.Context, query it.GetFieldMetadataQuery) (*it.GetFieldMetadataResult, error) {
	return this.fieldMetadataSvc.GetFieldMetadata(ctx, query)
}

func (this *FieldMetadataApplicationServiceImpl) SearchFieldMetadata(ctx corectx.Context, query it.SearchFieldMetadataQuery) (*it.SearchFieldMetadataResult, error) {
	return this.fieldMetadataSvc.SearchFieldMetadata(ctx, query)
}

func (this *FieldMetadataApplicationServiceImpl) UpdateFieldMetadata(ctx corectx.Context, cmd it.UpdateFieldMetadataCommand) (*it.UpdateFieldMetadataResult, error) {
	return this.fieldMetadataSvc.UpdateFieldMetadata(ctx, cmd)
}
