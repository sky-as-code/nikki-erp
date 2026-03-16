package dynamicentity

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"go.bryk.io/pkg/errors"
)

type DynamicEntityService struct {
	// modulePrefix limits this service to operate on entities of a specific module.
	modulePrefix string
	dbRepo       DbRepository
}

func (this *DynamicEntityService) Create(ctx context.Context, schemaName string, entity schema.DynamicEntity) (schema.DynamicEntity, error) {
	canonicalName := schema.CanonicalSchemaName(schemaName, this.modulePrefix)
	entitySchema := schema.GetSchema(canonicalName)
	if entitySchema == nil {
		return nil, errors.Errorf("schema '%s' not found", canonicalName)
	}

	validated, clientErrs := entitySchema.ValidateMap(map[string]any(entity), false)
	if clientErrs != nil && clientErrs.Count() > 0 {
		return nil, &fault.ClientError{
			Code:    "validation_error",
			Details: *clientErrs,
		}
	}

	collidingKeys, err := this.dbRepo.CheckUniqueCollisions(ctx, validated)
	if err != nil {
		return nil, err
	}
	if len(collidingKeys) > 0 {
		field := collidingKeys[0][0]
		return nil, &fault.ClientError{
			Code: "unique_constraint_violation",
			Details: fault.ClientErrors{fault.ClientErrorItem{
				Field:   field,
				Key:     "common.err_unique_constraint_violated",
				Message: "these unique constraints are violated {{.uniques}}",
				Type:    fault.ClientErrorTypeBusiness,
				Vars:    map[string]any{"uniques": collidingKeys},
			}},
		}
	}

	inserted, err := this.dbRepo.Insert(ctx, validated)
	if err != nil {
		return nil, err
	}
	return schema.DynamicEntity(inserted), nil
}
