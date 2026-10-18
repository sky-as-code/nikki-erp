// Package relationship declares the layers of the contacts_relationship resource: a typed link
// between two parties (employee of, subsidiary of, …).
package relationship

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type RelationshipRepository interface {
	composable.CrudRepository
}

type RelationshipDomainService interface {
	composable.CrudDomainService
}

type RelationshipApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateRelationshipCommand      = composable.CreateCommand
	UpdateRelationshipCommand      = composable.UpdateCommand
	DeleteRelationshipCommand      = composable.DeleteCommand
	SetRelationshipArchivedCommand = composable.SetArchivedCommand
	GetRelationshipByIdQuery       = composable.GetByIdQuery
	SearchRelationshipsQuery       = composable.SearchQuery
	RelationshipExistsQuery        = composable.ExistsQuery
)

type (
	CreateRelationshipResult      = composable.CreateResult
	UpdateRelationshipResult      = composable.MutateResult
	DeleteRelationshipResult      = composable.MutateResult
	SetRelationshipArchivedResult = composable.MutateResult
	GetRelationshipByIdResult     = composable.GetOneResult
	SearchRelationshipsResult     = composable.SearchResult
	RelationshipExistsResult      = composable.ExistsResult
)
