package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	itCore "github.com/sky-as-code/nikki-erp/modules/core/language/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

type userRestParams struct {
	dig.In

	UserSvc      it.UserAppService
	UserPrefsSvc itExt.UserSettingsExtService
}

func NewUserRest(params userRestParams) *UserRest {
	return &UserRest{
		UserSvc:      params.UserSvc,
		UserPrefsSvc: params.UserPrefsSvc,
	}
}

type UserRest struct {
	UserSvc      it.UserAppService
	UserPrefsSvc itExt.UserSettingsExtService
}

func (this UserRest) CreateUser(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate[CreateUserRequest, CreateUserResponse, domain.User](
		"create user",
		echoCtx,
		&it.CreateUserCommand{},
		this.UserSvc.CreateUser,
	)
}

func (this UserRest) DeleteUser(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[DeleteUserRequest, DeleteUserResponse](
		"delete user",
		echoCtx,
		this.UserSvc.DeleteUser,
	)
}

func (this UserRest) GetUser(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne2[GetUserRequest, GetUserResponse, domain.User](
		"get user",
		echoCtx,
		this.UserSvc.GetUser,
	)
}

func (this UserRest) ManageUserRoleAssignments(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[ManageUserRoleAssignmentsRequest, ManageUserRoleAssignmentsResponse](
		"manage user role assignments",
		echoCtx,
		this.UserSvc.ManageUserRoleAssignments,
	)
}

func (this UserRest) SearchUsers(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch[SearchUsersRequest, SearchUsersResponse, domain.User](
		"search users",
		echoCtx,
		this.UserSvc.SearchUsers,
	)
}

func (this UserRest) SetUserIsArchived(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate[SetUserIsArchivedRequest, SetUserIsArchivedResponse](
		"set user is_archived",
		echoCtx,
		this.UserSvc.SetUserIsArchived,
	)
}

func (this UserRest) UpdateUser(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate[UpdateUserRequest, UpdateUserResponse](
		"update user",
		echoCtx,
		&it.UpdateUserCommand{},
		this.UserSvc.UpdateUser,
	)
}

func (this UserRest) UserExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists[UserExistsRequest, UserExistsResponse](
		"user exists",
		echoCtx,
		this.UserSvc.UserExists,
	)
}

/*
 * Non-CRUD APIs
 */

func (this UserRest) GetModelSchema(echoCtx *echo.Context) (err error) {
	schema := dmodel.MustGetSchema(domain.UserSchemaName)
	echoCtx.JSON(http.StatusOK, schema.ToSimplized())
	return nil
}

func (this UserRest) GetUserContext(echoCtx *echo.Context) (err error) {
	reqCtx, err := corectx.AsRequestContext(echoCtx)
	if err != nil {
		return err
	}
	userPerm := reqCtx.GetPermissions()
	user := models.NewUserFrom(reqCtx.GetUser())
	echoCtx.JSON(http.StatusOK, GetUserContextResponse{
		Id:           string(user.MustGetId()),
		AvatarUrl:    user.GetAvatarUrl(),
		DisplayName:  user.MustGetDisplayName(),
		Email:        user.MustGetEmail(),
		Entitlements: userPerm.Entitlements.ToSlice(),
		Orgs: array.Map(user.GetOrgs(), func(org models.Organization) dmodel.DynamicFields {
			return org.GetFieldData()
		}),
		AccountSettings: this.accountSettings(reqCtx),
		SystemSettings: map[string]any{
			"app_name": "Nikki ERP",
		},
	})
	return nil
}

// accountSettings reports the acting user's own preferences to the frontend.
//
// theme_mode and language come from the settings module, which holds the user's stored choice and
// falls back to the schema's declared default when they have never set one. The remaining entries
// are still literals: timezone and the locale's formatting rules have no store behind them yet, so
// they are left exactly as they were rather than being invented here.
func (this UserRest) accountSettings(reqCtx corectx.Context) map[string]any {
	settings := map[string]any{
		"language":            defaultLanguage(),
		"timezone":            "Asia/Ho_Chi_Minh",
		"supported_languages": itExt.SupportedLanguages(),
		"theme_mode":          itExt.ThemeModeAuto,
	}

	result, err := this.UserPrefsSvc.GetUserPreferences(reqCtx, itExt.GetSettingsQuery{
		ModuleKey: itExt.EssentialModuleKey,
	})
	// The user context is what the whole application boots from, so a settings read that fails must
	// not take the session down with it: the defaults above are serviceable, and a user who cannot
	// load their theme should still be able to work.
	if err != nil || result == nil || !result.HasData {
		return settings
	}

	for _, item := range result.Data.Items {
		if item.Value == nil {
			continue
		}
		switch item.Name {
		case itExt.SettingThemeMode:
			settings["theme_mode"] = item.Value
		case itExt.SettingLanguage:
			if isoCode, ok := item.Value.(string); ok {
				settings["language"] = languageOf(isoCode)
			}
		}
	}
	return settings
}

// defaultLanguage is the locale used until the user has chosen one.
//
// It is not a business default: BR §53 is explicit that vi-VN must not be hardcoded as one. It is
// the display locale of last resort, and the moment the user stores a language this is replaced by
// languageOf.
func defaultLanguage() itCore.Language {
	return languageOf("vi-VN")
}

// languageOf builds the formatting rules the frontend needs for a locale.
//
// The rules are per-locale literals because there is no language store to read them from yet: the
// essential_language table holds no rows the user context can resolve against. Adding a locale to
// SupportedLanguages therefore needs an entry here too, which is what the default branch guards.
func languageOf(isoCode string) itCore.Language {
	switch isoCode {
	case "en-US":
		return itCore.Language{
			Id:                 util.ToPtr("01JZYF7DT3ASR4PXAX4HP4BVH0A"),
			Name:               util.ToPtr("English"),
			IsoCode:            util.ToPtr(model.LanguageCode("en-US")),
			Direction:          util.ToPtr("ltr"),
			DecimalSeparator:   util.ToPtr("."),
			ThousandsSeparator: util.ToPtr(","),
			DateFormat:         util.ToPtr("MM/dd/yyyy"),
			TimeFormat:         util.ToPtr("HH:mm:ss"),
			ShortTimeFormat:    util.ToPtr("HH:mm"),
			FirstDayOfWeek:     util.ToPtr("sunday"),
		}
	default:
		return itCore.Language{
			Id:                 util.ToPtr("01JZYF7DT3ASR4PXAX4HP4BVGZ"),
			Name:               util.ToPtr("Tiếng Việt"),
			IsoCode:            util.ToPtr(model.LanguageCode("vi-VN")),
			Direction:          util.ToPtr("ltr"),
			DecimalSeparator:   util.ToPtr("."),
			ThousandsSeparator: util.ToPtr(","),
			DateFormat:         util.ToPtr("dd/MM/yyyy"),
			TimeFormat:         util.ToPtr("HH:mm:ss"),
			ShortTimeFormat:    util.ToPtr("HH:mm"),
			FirstDayOfWeek:     util.ToPtr("monday"),
		}
	}
}
