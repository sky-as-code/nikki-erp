package v1

import (
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// GetSettingsRequest reads one module's settings for the acting owner at the level the route
// selects. The level is not a request field: it is fixed by which endpoint was called, so that a
// caller cannot reach the tenant level by sending a different string on an org-level route.
type GetSettingsRequest struct {
	ModuleKey string `param:"module_key"`
}

func (this GetSettingsRequest) ToQuery() it.GetSettingsQuery {
	return it.GetSettingsQuery{ModuleKey: this.ModuleKey}
}

// SetSettingsRequest carries only the items the caller changed. An absent item is left untouched
// rather than cleared, and — because writes are last-write-wins with no version check — sending an
// unchanged item is how a concurrent edit to a field this caller never touched gets clobbered.
type SetSettingsRequest struct {
	ModuleKey string              `param:"module_key"`
	Items     []SetSettingItemDto `json:"items"`
}

type SetSettingItemDto struct {
	Name  string `json:"name"`
	Value any    `json:"value"`

	// AllowOverride is accepted only from a tenant-level caller; see SetSettingItem. Absent leaves
	// the stored policy untouched, so an org or user PATCH never has to send it.
	AllowOverride *bool `json:"allow_override,omitempty"`
}

func (this SetSettingsRequest) ToCommand() it.SetSettingsCommand {
	items := make([]it.SetSettingItem, 0, len(this.Items))
	for _, item := range this.Items {
		items = append(items, it.SetSettingItem{
			Name:          item.Name,
			Value:         item.Value,
			AllowOverride: item.AllowOverride,
		})
	}
	return it.SetSettingsCommand{
		ModuleKey: this.ModuleKey,
		Items:     items,
	}
}

// SettingItemResponse is one setting as the UI sees it: the value together with enough of its
// declaration to render and validate a control without re-reading the schema.
type SettingItemResponse struct {
	Name          string `json:"name"`
	Level         string `json:"level"`
	Value         any    `json:"value"`
	HasValue      bool   `json:"has_value"`
	AllowOverride bool   `json:"allow_override"`
	Editable      bool   `json:"editable"`
	Field         any    `json:"field,omitempty"`
}

type GetSettingsResponse struct {
	ModuleKey string `json:"module_key"`
	// OwnerType is the scope the caller is reading as. See GetSettingsResultData.OwnerType: a
	// client cannot infer it from `editable`, which is true at one's own level for every actor.
	OwnerType string                `json:"owner_type"`
	Items     []SettingItemResponse `json:"items"`
}

func NewGetSettingsResponse(data it.GetSettingsResultData) GetSettingsResponse {
	items := make([]SettingItemResponse, 0, len(data.Items))
	for _, item := range data.Items {
		items = append(items, newSettingItemResponse(item))
	}
	return GetSettingsResponse{
		ModuleKey: data.ModuleKey,
		OwnerType: data.OwnerType,
		Items:     items,
	}
}

func newSettingItemResponse(item it.SettingItem) SettingItemResponse {
	response := SettingItemResponse{
		Name:          item.Name,
		Level:         item.Level,
		Value:         item.Value,
		HasValue:      item.HasValue,
		AllowOverride: item.AllowOverride,
		Editable:      item.Editable,
	}

	// The field is simplized rather than sent raw: ModelField holds unexported state and builder
	// wiring that has no meaning to a client, and ToSimplized is the same shape the dynamic schema
	// endpoints already return, so the frontend renders both from one code path.
	if item.Field != nil {
		response.Field = item.Field.ToSimplized()
	}
	return response
}

type SetSettingsResponse struct {
	// Updated counts the rows written, including the children an enforced tenant setting fanned
	// out onto, so a tenant admin can see that the change reached more than one row.
	Updated int `json:"updated"`
}

func NewSetSettingsResponse(data it.SetSettingsResultData) SetSettingsResponse {
	return SetSettingsResponse{Updated: data.Updated}
}
