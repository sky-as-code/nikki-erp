package schema

import (
	"context"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicentity/model"
)

// EntityRepository defines the interface for entity repository operations.
// Methods match the signatures from DynamicRepositoryBase.
type EntityRepository interface {
	// CreateM creates a new record from an EntityMap.
	// Returns an EntityMap containing only Primary & Tenant Keys.
	CreateM(ctx context.Context, data dmodel.EntityMap) (dmodel.EntityMap, error)

	// CreateS creates a new record from a struct.
	// Uses StructToEntityMap to convert the struct to EntityMap, then invokes CreateM.
	CreateS(ctx context.Context, data any) (dmodel.EntityMap, error)

	// CreateBulkM creates multiple records from EntityMaps.
	CreateBulkM(ctx context.Context, data []dmodel.EntityMap) ([]dmodel.EntityMap, error)

	// CreateBulkS creates multiple records from structs.
	// Uses StructToEntityMap to convert each struct to EntityMap, then invokes CreateBulkM.
	CreateBulkS(ctx context.Context, data []any) ([]dmodel.EntityMap, error)

	// UpdateM updates a record by primary key from an EntityMap.
	UpdateM(ctx context.Context, data dmodel.EntityMap) error

	// UpdateS updates a record by primary key from a struct.
	// Uses StructToEntityMap to convert the struct to EntityMap, then invokes UpdateM.
	UpdateS(ctx context.Context, data any) error

	// Delete deletes records matching the filters.
	// Returns the number of rows deleted.
	Delete(ctx context.Context, filters map[string]string) (int64, error)

	// Exists checks if a record exists matching the filters.
	Exists(ctx context.Context, filters map[string]string) (bool, error)
}
