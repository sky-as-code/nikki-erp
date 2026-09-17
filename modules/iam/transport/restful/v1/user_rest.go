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

	UserSvc     it.UserAppService
	SettingsSvc itExt.EffectiveSettingsExtService
	LanguageSvc itExt.LanguageExtService
	CurrencySvc itExt.CurrencyExtService
}

func NewUserRest(params userRestParams) *UserRest {
	return &UserRest{
		UserSvc:     params.UserSvc,
		SettingsSvc: params.SettingsSvc,
		LanguageSvc: params.LanguageSvc,
		CurrencySvc: params.CurrencySvc,
	}
}

type UserRest struct {
	UserSvc     it.UserAppService
	SettingsSvc itExt.EffectiveSettingsExtService
	LanguageSvc itExt.LanguageExtService
	CurrencySvc itExt.CurrencyExtService
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

		IsOwner: userPerm.IsOwner,
		UserOrgIds: array.Map(userPerm.UserOrgIds.ToSlice(), func(id model.Id) string {
			return string(id)
		}),
		OrgUnitId:    idToStringPtr(userPerm.OrgUnitId),
		OrgUnitOrgId: idToStringPtr(userPerm.OrgUnitOrgId),

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

// idToStringPtr carries an optional id across to the wire, keeping absence as null rather than
// collapsing it to an empty string: the frontend's evaluator treats "no org unit" and "an org
// unit whose id is blank" differently.
func idToStringPtr(id *model.Id) *string {
	if id == nil {
		return nil
	}
	return util.ToPtr(string(*id))
}

// accountSettings reports the settings the acting user's session runs on.
//
// Every value is read through the Settings module's consolidated read rather than level by level:
// one call answers "what applies to this user", already folded across tenant, org and user. This
// handler then resolves the two references in that answer -- a language iso code and a currency
// code -- against Essential's catalogues, so the frontend receives records rather than codes it
// would have to look up itself.
//
// It stays fail-soft throughout. The user context is what the whole application boots from, so a
// settings or catalogue read that fails degrades the entry rather than the session: a user who
// cannot load their currency should still be able to work.
func (this UserRest) accountSettings(reqCtx corectx.Context) map[string]any {
	settings := map[string]any{
		"language":            fallbackLanguage(),
		"timezone":            "Asia/Ho_Chi_Minh",
		"supported_languages": itExt.SupportedLanguages(),
		"theme_mode":          itExt.ThemeModeAuto,
	}

	result, err := this.SettingsSvc.GetEffectiveSettings(reqCtx, itExt.GetEffectiveSettingsQuery{})
	if err != nil || result == nil || !result.HasData {
		return settings
	}
	values := result.Data.Values

	if mode, ok := values[essentialKey(itExt.SettingThemeMode)].(string); ok {
		settings["theme_mode"] = mode
	}
	if language := this.resolveLanguage(reqCtx, values); language != nil {
		settings["language"] = *language
	}
	if currency := this.resolveCurrency(reqCtx, values); currency != nil {
		settings["currency"] = currency
	}
	return settings
}

// essentialKey names a setting the way GetEffectiveSettings keys it.
func essentialKey(name string) string {
	return itExt.EssentialModuleKey + "." + name
}

// resolveLanguage reads the acting user's display language out of the effective settings and
// resolves it against essential_languages.
//
// The fallback chain -- the user's own language, then the organization's system_language -- is
// deliberately the same one essential/app.NewUserLocaleResolver walks, so that the language a list
// is sorted in is the language the frontend formats it in. Keep the two in step; they are separate
// because that resolver yields a language code for the search layer while this needs the whole
// formatting record.
func (this UserRest) resolveLanguage(ctx corectx.Context, values map[string]any) *itCore.Language {
	for _, key := range []string{
		essentialKey(itExt.SettingLanguage),
		essentialKey(itExt.SettingSystemLanguage),
	} {
		isoCode, ok := values[key].(string)
		if !ok || !array.Contains(itExt.SupportedLanguages(), isoCode) {
			continue
		}
		if language := this.loadLanguage(ctx, isoCode); language != nil {
			return language
		}
	}
	return nil
}

func (this UserRest) loadLanguage(ctx corectx.Context, isoCode string) *itCore.Language {
	found, err := this.LanguageSvc.GetLanguageByIsoCode(ctx, itExt.GetLanguageByIsoCodeQuery{
		IsoCode: isoCode,
	})
	if err != nil || found == nil || !found.HasData {
		return nil
	}
	record := found.Data
	return &itCore.Language{
		Id:                 record.GetId(),
		Name:               record.GetName(),
		IsoCode:            util.ToPtr(model.LanguageCode(util.ValueOrZeroOf(record.GetIsoCode()))),
		Direction:          record.GetDirection(),
		DecimalSeparator:   record.GetDecimalSeparator(),
		ThousandsSeparator: record.GetThousandsSeparator(),
		DateFormat:         record.GetDateFormat(),
		TimeFormat:         record.GetTimeFormat(),
		ShortTimeFormat:    record.GetShortTimeFormat(),
		FirstDayOfWeek:     record.GetFirstDayOfWeek(),
	}
}

// resolveCurrency resolves the organization's default_currency, which stores an alphabetic code.
//
// An absent or unresolvable currency omits the key rather than guessing one. There is no safe
// default currency: a guessed one silently reinterprets every amount on screen.
func (this UserRest) resolveCurrency(ctx corectx.Context, values map[string]any) map[string]any {
	code, ok := values[essentialKey(itExt.SettingDefaultCurrency)].(string)
	if !ok || code == "" {
		return nil
	}
	found, err := this.CurrencySvc.GetCurrencyByCode(ctx, itExt.GetCurrencyByCodeQuery{Code: code})
	if err != nil || found == nil || !found.HasData {
		return nil
	}
	return map[string]any{
		"code":   found.Data.Code,
		"symbol": found.Data.Symbol,
		// The currency catalogue seeds no symbols -- a wrong symbol on money is worse than none --
		// so the frontend renders the code whenever symbol is empty.
		"decimal_places": found.Data.DecimalPlaces,
	}
}

// fallbackLanguage is the display language of last resort, used only when the settings read fails
// outright or names a language essential_languages cannot resolve.
//
// It is a literal because the response must always carry a language: the frontend formats every
// number and date from it, and a session that booted without one would render nothing at all. It is
// en-US rather than any business-chosen language -- BR §53 is explicit that vi-VN must not be
// hardcoded as a default -- and it is replaced the moment a real record resolves.
func fallbackLanguage() itCore.Language {
	return itCore.Language{
		Name:               util.ToPtr("English"),
		IsoCode:            util.ToPtr(model.LanguageCode(model.DefaultLanguageCode)),
		Direction:          util.ToPtr("ltr"),
		DecimalSeparator:   util.ToPtr("."),
		ThousandsSeparator: util.ToPtr(","),
		DateFormat:         util.ToPtr("MM/dd/yyyy"),
		TimeFormat:         util.ToPtr("HH:mm:ss"),
		ShortTimeFormat:    util.ToPtr("HH:mm"),
		FirstDayOfWeek:     util.ToPtr("sunday"),
	}
}
