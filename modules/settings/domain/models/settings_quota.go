package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SettingsQuotaSchemaName = "settings_quota"

	SettingsQuotaFieldId                       = basemodel.FieldId
	SettingsQuotaFieldName                     = "name"
	SettingsQuotaFieldOwnerType                = "owner_type"
	SettingsQuotaFieldOwnerId                  = "owner_id"
	SettingsQuotaFieldItStorageCapacityMibytes = "itstorage_capacity_mibytes"
)

//go:embed settings_quota.json
var settingsQuotaSchemaJson string

func SettingsQuotaSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(settingsQuotaSchemaJson)
}

// SettingsQuota caps a resource, named by Name, for one tenant, organization or user.
type SettingsQuota struct {
	basemodel.DynamicModelBase
}

func NewSettingsQuotaFrom(src dmodel.DynamicFields) *SettingsQuota {
	return &SettingsQuota{DynamicModelBase: basemodel.NewDynamicModel(src)}
}

func (this SettingsQuota) GetName() *string {
	return this.GetFieldData().GetString(SettingsQuotaFieldName)
}

func (this SettingsQuota) GetOwnerType() *string {
	return this.GetFieldData().GetString(SettingsQuotaFieldOwnerType)
}

func (this SettingsQuota) GetOwnerId() *model.Id {
	return this.GetFieldData().GetModelId(SettingsQuotaFieldOwnerId)
}

// GetItStorageCapacityMibytes is nil when the row sets no storage cap.
func (this SettingsQuota) GetItStorageCapacityMibytes() *int64 {
	capacity, ok := util.AnyToInt64(this.GetFieldData().GetAny(SettingsQuotaFieldItStorageCapacityMibytes))
	if !ok {
		return nil
	}
	return &capacity
}
