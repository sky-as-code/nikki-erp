// Package relationship declares the layers of the contacts_relationship resource: a typed link
// between two parties (employee of, subsidiary of, …).
package relationship

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type RelationshipRepository interface {
	dynamicresource.CrudRepository
}

type RelationshipDomainService interface {
	dynamicresource.CrudDomainService
}

type RelationshipApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateRelationshipCommand      = dynamicresource.CreateCommand
	UpdateRelationshipCommand      = dynamicresource.UpdateCommand
	DeleteRelationshipCommand      = dynamicresource.DeleteCommand
	SetRelationshipArchivedCommand = dynamicresource.SetArchivedCommand
	GetRelationshipByIdQuery       = dynamicresource.GetByIdQuery
	SearchRelationshipsQuery       = dynamicresource.SearchQuery
	RelationshipExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateRelationshipResult      = dynamicresource.CreateResult
	UpdateRelationshipResult      = dynamicresource.MutateResult
	DeleteRelationshipResult      = dynamicresource.MutateResult
	SetRelationshipArchivedResult = dynamicresource.MutateResult
	GetRelationshipByIdResult     = dynamicresource.GetOneResult
	SearchRelationshipsResult     = dynamicresource.SearchResult
	RelationshipExistsResult      = dynamicresource.ExistsResult
)
