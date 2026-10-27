package repository

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	itService "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/service"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

// NewServiceCrudRepository views the user repository through the iam_service schema. Every query
// still runs on the iam_user schema (the table's owner, so `type` and the junctions resolve), while
// the generic crud helpers - which validate a write and expand `*` against GetBaseRepo().Schema() -
// see iam_service, so a service is validated without the human-only fields.
func NewServiceCrudRepository(
	userRepo itUser.UserCrudRepository, serviceSchema *dmodel.ModelSchema,
) itService.ServiceCrudRepository {
	return &ServiceCrudRepositoryImpl{
		UserCrudRepository: userRepo,
		schema:             serviceSchema,
		baseRepo:           schemaView{BaseDynamicRepository: userRepo.GetBaseRepo(), schema: serviceSchema},
	}
}

type ServiceCrudRepositoryImpl struct {
	itUser.UserCrudRepository
	schema   *dmodel.ModelSchema
	baseRepo dyn.BaseDynamicRepository
}

func (this *ServiceCrudRepositoryImpl) Schema() *dmodel.ModelSchema {
	return this.schema
}

func (this *ServiceCrudRepositoryImpl) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.baseRepo
}

// schemaView reports another schema than the one its repository queries with.
type schemaView struct {
	dyn.BaseDynamicRepository
	schema *dmodel.ModelSchema
}

func (this schemaView) Schema() *dmodel.ModelSchema {
	return this.schema
}
