package module

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
)

type CreateModuleCommand struct {
	Label   model.LangJson `json:"label"`
	Name    string         `json:"name"`
	Version string         `json:"version"`
}

type CreateModuleResult = GetModuleResult

type CreateBulkModulesCommand struct {
	Modules []CreateModuleCommand `json:"modules"`
}

type CreateBulkModulesResult = crud.OpResult[[]*domain.ModuleMetadata]

type UpdateModuleCommand struct {
	Id         model.Id        `param:"id" json:"id"`
	Name       *string         `json:"name"`
	Label      *model.LangJson `json:"label"`
	IsOrphaned *bool           `json:"isOrphaned"`
	Version    *string         `json:"version"`
}

type UpdateModuleResult = GetModuleResult

type UpdateBulkModulesCommand struct {
	Modules []UpdateModuleCommand `json:"modules"`
}

type UpdateBulkModulesResult = crud.OpResult[[]*domain.ModuleMetadata]

type DeleteModuleCommand = GetModuleByIdQuery

type DeleteModuleResultData struct {
	Id           *model.Id `json:"id"`
	DeletedAt    time.Time `json:"deletedAt"`
	DeletedCount int       `json:"deletedCount"`
}

type DeleteModuleResult = crud.DeletionResult

type ModuleExistsQuery = GetModuleByIdQuery
type ModuleExistsByNameQuery = GetModuleByNameQuery

type ModuleExistsResult = crud.ExistsResult

type GetModuleByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (this GetModuleByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetModuleByNameQuery struct {
	Name string `json:"name"`
}

func (this ModuleExistsByNameQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.NotEmpty,
			val.Length(1, model.MODEL_RULE_TINY_NAME_LENGTH),
		),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetModuleResult = crud.OpResult[*domain.ModuleMetadata]

type ListModulesQuery struct {
}

func (this ListModulesQuery) Validate() ft.ValidationErrors {
	return nil
}

type ListModulesResult = crud.OpResult[[]domain.ModuleMetadata]

type SearchModulesQuery struct {
	crud.SearchQuery
}

func (this SearchModulesQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchModulesResultData = crud.PagedResult[domain.ModuleMetadata]
type SearchModulesResult = crud.OpResult[*SearchModulesResultData]
