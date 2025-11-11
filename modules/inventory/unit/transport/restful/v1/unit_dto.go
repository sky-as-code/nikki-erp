package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/unit/interfaces"
)

type UnitDto struct {
	Id         string         `json:"id"`
	BaseUnit   *string        `json:"baseUnit,omitempty"`
	CategoryId *string        `json:"categoryId,omitempty"`
	CreatedAt  int64          `json:"createdAt"`
	Etag       string         `json:"etag"`
	Multiplier *float64       `json:"multiplier,omitempty"`
	Name       model.LangJson `json:"name"`
	OrgId      *string        `json:"orgId,omitempty"`
	UpdatedAt  *int64         `json:"updatedAt,omitempty"`
	Status     *string        `json:"status,omitempty"`
}

func (this *UnitDto) FromUnit(u it.Unit) {
	model.MustCopy(u.AuditableBase, this)
	model.MustCopy(u.ModelBase, this)
	model.MustCopy(u, this)
}

type CreateUnitRequest = it.CreateUnitCommand
type CreateUnitResponse = httpserver.RestCreateResponse

type UpdateUnitRequest = it.UpdateUnitCommand
type UpdateUnitResponse = httpserver.RestUpdateResponse

type DeleteUnitRequest = it.DeleteUnitCommand
type DeleteUnitResponse = httpserver.RestDeleteResponse

type GetUnitByIdRequest = it.GetUnitByIdQuery
type GetUnitByIdResponse = UnitDto

type SearchUnitsRequest = it.SearchUnitsQuery

type SearchUnitsResponse httpserver.RestSearchResponse[UnitDto]

func (this *SearchUnitsResponse) FromResult(result *it.SearchUnitsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(u it.Unit) UnitDto {
		item := UnitDto{}
		item.FromUnit(u)
		return item
	})
}
