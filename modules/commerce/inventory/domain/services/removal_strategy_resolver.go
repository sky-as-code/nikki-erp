package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
)

// maxLocationDepth caps the upward walk; a parent cycle the engine failed to prevent must not
// loop forever.
const maxLocationDepth = 64

// removalStrategyReader is what the resolver reads, split out so its order can be tested without
// a database.
type removalStrategyReader interface {
	CategoryStrategy(ctx corectx.Context, categoryId string) (string, error)
	LocationStrategy(ctx corectx.Context, locationId string) (strategy string, parentId string, err error)
}

// ResolveRemovalStrategy picks the order goods leave sourceLocationId in: the product category's
// force_removal_strategy, then the location's own removal_strategy, then the nearest ancestor's,
// then FIFO. Either id may be empty.
func ResolveRemovalStrategy(ctx corectx.Context, productCategoryId string, sourceLocationId string) (string, error) {
	return resolveRemovalStrategy(ctx, repoRemovalStrategyReader{}, productCategoryId, sourceLocationId)
}

func resolveRemovalStrategy(
	ctx corectx.Context, reader removalStrategyReader, categoryId string, locationId string,
) (string, error) {
	if categoryId != "" {
		forced, err := reader.CategoryStrategy(ctx, categoryId)
		if err != nil || forced != "" {
			return forced, err
		}
	}
	visited := map[string]bool{}
	current := locationId
	for current != "" && !visited[current] && len(visited) < maxLocationDepth {
		visited[current] = true
		strategy, parentId, err := reader.LocationStrategy(ctx, current)
		if err != nil || strategy != "" {
			return strategy, err
		}
		current = parentId
	}
	return models.InventoryLocationRemovalStrategyFifo, nil
}

// resolveVariantRemovalStrategy resolves for a variant, reading its category through the
// variant's related field.
func resolveVariantRemovalStrategy(ctx corectx.Context, variantId string, locationId string) (string, error) {
	categoryId, err := readStringField(ctx, models.ProductVariantSchemaName, variantId,
		models.ProductVariantFieldTemplateCategoryId)
	if err != nil {
		return "", err
	}
	return ResolveRemovalStrategy(ctx, categoryId, locationId)
}

// lockQuantsInRemovalOrder locks a variant's balances in a location and returns them in the order
// the resolved removal strategy consumes them.
func lockQuantsInRemovalOrder(
	ctx corectx.Context, repo dyn.BaseDynamicRepository, key QuantLockKey,
) ([]LockedQuant, error) {
	locked, err := LockQuantsForUpdate(ctx, repo, key)
	if err != nil || len(locked) < 2 {
		return locked, err
	}
	strategy, err := resolveVariantRemovalStrategy(ctx, string(key.ProductVariantId), string(key.LocationId))
	if err != nil {
		return nil, err
	}
	return OrderQuantsForRemoval(locked, strategy), nil
}

type repoRemovalStrategyReader struct{}

func (this repoRemovalStrategyReader) CategoryStrategy(ctx corectx.Context, categoryId string) (string, error) {
	return readStringField(ctx, models.ProductCategorySchemaName, categoryId,
		models.ProductCategoryFieldForceRemovalStrategy)
}

func (this repoRemovalStrategyReader) LocationStrategy(
	ctx corectx.Context, locationId string,
) (string, string, error) {
	record, err := readFields(ctx, models.InventoryLocationSchemaName, locationId,
		models.InventoryLocationFieldRemovalStrategy, models.InventoryLocationFieldParentLocationId)
	if err != nil || record == nil {
		return "", "", err
	}
	return derefString(record.GetString(models.InventoryLocationFieldRemovalStrategy)),
		derefString(record.GetString(models.InventoryLocationFieldParentLocationId)), nil
}

func readStringField(ctx corectx.Context, schemaName string, id string, field string) (string, error) {
	record, err := readFields(ctx, schemaName, id, field)
	if err != nil || record == nil {
		return "", err
	}
	return derefString(record.GetString(field)), nil
}

// readFields reads the named columns of one record; a missing record is nil, not an error.
func readFields(ctx corectx.Context, schemaName string, id string, fields ...string) (dmodel.DynamicFields, error) {
	engine, err := repoFor(schemaName)
	if err != nil {
		return nil, err
	}
	found, err := engine.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{basemodel.FieldId: id},
		Fields: append([]string{basemodel.FieldId}, fields...),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "readFields: %s '%s'", schemaName, id)
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data, nil
}
