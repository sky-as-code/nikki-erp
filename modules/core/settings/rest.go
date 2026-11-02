package settings

import (
	"github.com/labstack/echo/v5"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

// UpdateSettingsRequest is the body of every module's `PATCH /v1/{module_key}/settings/:level`.
// One shape for every module is what lets a frontend settings widget derive the endpoint from its
// module key alone.
//
// It carries only the items the caller changed. An absent item is left untouched rather than
// cleared, and — because writes are last-write-wins with no version check — sending an unchanged
// item is how a concurrent edit to a field this caller never touched gets clobbered.
type UpdateSettingsRequest struct {
	Level string                 `param:"level" json:"-"`
	Items []UpdateSettingItemDto `json:"items"`
	// OrgId names the org an org-level write acts for, as on every org-scoped resource.
	OrgId *string `json:"org_id,omitempty"`
}

type UpdateSettingItemDto struct {
	Name  string `json:"name"`
	Value any    `json:"value"`

	// AllowOverride is accepted only from a tenant-level caller; see SetSettingItem. Absent leaves
	// the stored policy untouched, so an org or user PATCH never has to send it.
	AllowOverride *bool `json:"allow_override,omitempty"`
}

func (this UpdateSettingsRequest) ToCommand() UpdateSettingsCommand {
	items := make([]SetSettingItem, 0, len(this.Items))
	for _, item := range this.Items {
		items = append(items, SetSettingItem{Name: item.Name, Value: item.Value, AllowOverride: item.AllowOverride})
	}
	return UpdateSettingsCommand{Level: this.Level, Items: items, OrgId: OptionalOrgId(this.OrgId)}
}

// OptionalOrgId reads an optional org_id parameter; absent or blank both mean "not named".
func OptionalOrgId(raw *string) *model.Id {
	if raw == nil || *raw == "" {
		return nil
	}
	orgId := model.Id(*raw)
	return &orgId
}

type UpdateSettingsResponse struct {
	// Updated counts the rows written, including the children an enforced tenant setting fanned
	// out onto, so a tenant admin can see that the change reached more than one row.
	Updated int `json:"updated"`
}

func NewUpdateSettingsResponse(data SetSettingsResultData) UpdateSettingsResponse {
	return UpdateSettingsResponse{Updated: data.Updated}
}

// UpdateSettingsRoutePath is mounted under the module's own `/v1/{module_key}` group.
const UpdateSettingsRoutePath = "/settings/:level"

// UpdateSettingsHandler is a module's settings update route, mounted as
// `route.PATCH(UpdateSettingsRoutePath, UpdateSettingsHandler(svc.UpdateSettings), m.UserSmokeAuthz())`.
func UpdateSettingsHandler(
	updateFn func(corectx.Context, UpdateSettingsCommand) (*SetSettingsResult, error),
) echo.HandlerFunc {
	return func(echoCtx *echo.Context) (err error) {
		defer func() {
			if e := ft.RecoverPanicFailedTo(recover(), "handle REST update settings"); e != nil {
				err = e
			}
		}()
		return httpserver.ServeRequest2(
			echoCtx, updateFn, UpdateSettingsRequest.ToCommand, NewUpdateSettingsResponse, httpserver.JsonOk,
		)
	}
}
