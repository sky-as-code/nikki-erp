package enum

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type EnumRepository interface {
	composable.CrudRepository
}

type EnumDomainService interface {
	composable.CrudDomainService
}

type EnumApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateEnumCommand = composable.CreateCommand
	UpdateEnumCommand = composable.UpdateCommand
	DeleteEnumCommand = composable.DeleteCommand
	GetEnumByIdQuery  = composable.GetByIdQuery
	SearchEnumsQuery  = composable.SearchQuery
	EnumExistsQuery   = composable.ExistsQuery
)

type (
	CreateEnumResult  = composable.CreateResult
	UpdateEnumResult  = composable.MutateResult
	DeleteEnumResult  = composable.MutateResult
	GetEnumByIdResult = composable.GetOneResult
	SearchEnumsResult = composable.SearchResult
	EnumExistsResult  = composable.ExistsResult
)
