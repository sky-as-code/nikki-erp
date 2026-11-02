package enum

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type EnumRepository interface {
	dynamicresource.CrudRepository
}

type EnumDomainService interface {
	dynamicresource.CrudDomainService
}

type EnumApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateEnumCommand = dynamicresource.CreateCommand
	UpdateEnumCommand = dynamicresource.UpdateCommand
	DeleteEnumCommand = dynamicresource.DeleteCommand
	GetEnumByIdQuery  = dynamicresource.GetByIdQuery
	SearchEnumsQuery  = dynamicresource.SearchQuery
	EnumExistsQuery   = dynamicresource.ExistsQuery
)

type (
	CreateEnumResult  = dynamicresource.CreateResult
	UpdateEnumResult  = dynamicresource.MutateResult
	DeleteEnumResult  = dynamicresource.MutateResult
	GetEnumByIdResult = dynamicresource.GetOneResult
	SearchEnumsResult = dynamicresource.SearchResult
	EnumExistsResult  = dynamicresource.ExistsResult
)
