// Package space declares the layers of the itstorage_space resource: the system-managed tree that
// files are stored in.
//
// Over REST the resource is read-only. Spaces are created by the org lifecycle and by seeds.
package space

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
)

// SpaceRepository adds the set-based statements the tree needs. Each joins the transaction the
// context carries and confines itself to the acting tenant.
type SpaceRepository interface {
	composable.CrudRepository

	// AddSize adds delta to the space and every ancestor in one statement, never below zero.
	AddSize(ctx corectx.Context, spaceId model.Id, delta int64) error

	// RewriteDescendantPaths replaces the oldPath prefix of every descendant with newPath, in one
	// UPDATE, and returns the rows touched. The renamed space itself is not included.
	RewriteDescendantPaths(ctx corectx.Context, oldPath string, newPath string) (int, error)

	// ListSubtreeIds returns the space at path and every descendant, deepest first.
	ListSubtreeIds(ctx corectx.Context, path string) ([]model.Id, error)

	// DeleteSpaces deletes the given spaces in order; children must precede their parents.
	DeleteSpaces(ctx corectx.Context, spaceIds []model.Id) (int, error)
}

type SpaceDomainService interface {
	composable.CrudDomainService

	// FindSpace returns nil when no space has the id.
	FindSpace(ctx corectx.Context, spaceId model.Id) (*models.Space, error)
	// FindTenantRootSpace returns nil when the acting tenant has no root space, which is always
	// the case in a binary without tenants.
	FindTenantRootSpace(ctx corectx.Context) (*models.Space, error)
	// FindOrgSpace returns nil when the org has no space yet.
	FindOrgSpace(ctx corectx.Context, orgId model.Id) (*models.Space, error)

	AddSize(ctx corectx.Context, spaceId model.Id, delta int64) error
}

type SpaceApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateSpaceCommand = composable.CreateCommand
	UpdateSpaceCommand = composable.UpdateCommand
	GetSpaceByIdQuery  = composable.GetByIdQuery
	SearchSpacesQuery  = composable.SearchQuery
	CreateSpaceResult  = composable.CreateResult
	MutateSpaceResult  = composable.MutateResult
	GetSpaceByIdResult = composable.GetOneResult
	SearchSpacesResult = composable.SearchResult
)
