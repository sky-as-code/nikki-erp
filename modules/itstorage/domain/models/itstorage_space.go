package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	c "github.com/sky-as-code/nikki-erp/modules/itstorage/constants"
)

const (
	SpaceSchemaName = c.ResourceItStorageSpace

	SpaceFieldId            = basemodel.FieldId
	SpaceFieldEtag          = basemodel.FieldEtag
	SpaceFieldName          = "name"
	SpaceFieldParentId      = "parent_id"
	SpaceFieldPath          = "path"
	SpaceFieldSizeBytes     = "size_bytes"
	SpaceFieldOwnerType     = "owner_type"
	SpaceFieldOwnerId       = "owner_id"
	SpaceFieldManagedByType = "managed_by_type"
	SpaceFieldManagedById   = "managed_by_id"
)

//go:embed itstorage_space.json
var spaceSchemaJson string

func SpaceSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(spaceSchemaJson)
}

// Space is a node of the storage tree. Its path is materialised, so a file path is resolvable
// without walking the tree.
type Space struct {
	basemodel.DynamicModelBase
}

func NewSpaceFrom(src dmodel.DynamicFields) *Space {
	return &Space{DynamicModelBase: basemodel.NewDynamicModel(src)}
}

func (this Space) GetName() *string {
	return this.GetFieldData().GetString(SpaceFieldName)
}

func (this Space) GetParentId() *model.Id {
	return this.GetFieldData().GetModelId(SpaceFieldParentId)
}

func (this Space) GetPath() string {
	return util.ValueOrZeroOf(this.GetFieldData().GetString(SpaceFieldPath))
}

func (this Space) GetSizeBytes() int64 {
	size, _ := util.AnyToInt64(this.GetFieldData().GetAny(SpaceFieldSizeBytes))
	return size
}

func (this Space) GetOwnerType() string {
	return util.ValueOrZeroOf(this.GetFieldData().GetString(SpaceFieldOwnerType))
}

func (this Space) GetOwnerId() *model.Id {
	return this.GetFieldData().GetModelId(SpaceFieldOwnerId)
}

// ContainsPath reports whether path is this space's own path or lies below it.
func (this Space) ContainsPath(path string) bool {
	own := this.GetPath()
	return own != "" && (path == own || IsUnderPath(path, own))
}
