// Package consumption declares the layers of the itstorage_consumption resource: the bytes each
// tenant, organization and user stores, and the bytes reserved by uploads still in flight.
//
// Over REST the resource is read-only. The upload, duplicate and delete paths keep it.
package consumption

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
)

// ConsumptionRepository adds the conditional counters. Each statement joins the transaction the
// context carries and confines itself to the acting tenant.
type ConsumptionRepository interface {
	composable.CrudRepository

	// Ensure creates the missing rows, seeded with InitialUsedBytes; existing rows are untouched.
	Ensure(ctx corectx.Context, owners []models.ConsumptionOwner) error
	// Reserve adds bytes to reserved_bytes. With capBytes, only while used + reserved + bytes stays
	// within it: false means the cap would be exceeded and nothing changed.
	Reserve(ctx corectx.Context, owner models.ConsumptionOwner, bytes int64, capBytes *int64) (bool, error)
	// Commit moves bytes from reserved_bytes to used_bytes.
	Commit(ctx corectx.Context, owners []models.ConsumptionOwner, bytes int64) error
	// Release takes bytes off reserved_bytes.
	Release(ctx corectx.Context, owners []models.ConsumptionOwner, bytes int64) error
	// Subtract takes bytes off used_bytes.
	Subtract(ctx corectx.Context, owners []models.ConsumptionOwner, bytes int64) error
	// SubtractEach takes each owner's own amount off its used_bytes.
	SubtractEach(ctx corectx.Context, ownerType string, bytesByOwner map[model.Id]int64) error
	// DeleteOwner removes an owner's row, for an organization that is cleaned up.
	DeleteOwner(ctx corectx.Context, owner models.ConsumptionOwner) error
}

// Reservation is the bytes one upload or duplicate holds on its owners until it commits or fails.
// CapBytes applies to Owners[0], the tenant; the other owners are counted without a cap.
type Reservation struct {
	Owners   []models.ConsumptionOwner
	Bytes    int64
	CapBytes *int64
}

type ConsumptionDomainService interface {
	composable.CrudDomainService

	// Reserve ensures the owners' rows and reserves on all of them in one short transaction of its
	// own, so no lock is held while the content streams. false means the cap refused it.
	Reserve(ctx corectx.Context, reservation Reservation) (bool, error)
	// Commit runs in the transaction that records the file.
	Commit(ctx corectx.Context, reservation Reservation) error
	// Release gives a failed upload's reservation back. It runs detached from the request, never
	// fails the caller, and logs what it could not release. Inside the caller's transaction it does
	// nothing: the reservation was made in that transaction and rolls back with it.
	Release(ctx corectx.Context, reservation Reservation)
	// Subtract runs in the transaction that deletes the file.
	Subtract(ctx corectx.Context, owners []models.ConsumptionOwner, bytes int64) error
	// SubtractEach and DeleteOwner serve the org cleanup.
	SubtractEach(ctx corectx.Context, ownerType string, bytesByOwner map[model.Id]int64) error
	DeleteOwner(ctx corectx.Context, owner models.ConsumptionOwner) error
}

type ConsumptionApplicationService interface {
	composable.CrudApplicationService
}

type (
	GetConsumptionByIdQuery = composable.GetByIdQuery
	SearchConsumptionsQuery = composable.SearchQuery
)
