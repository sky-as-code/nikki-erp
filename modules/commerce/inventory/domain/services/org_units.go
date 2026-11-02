package services

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	invconstants "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/constants"
	invsettings "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// OrgUnits maps each unit setting (default_weight_uom_id, default_length_uom_id) to the unit an
// organization's template weight and dimensions are expressed in. A unit the organization does not
// have is absent.
type OrgUnits map[string]*itExt.GetUomResultData

// OrgUnitResolver answers which units an organization measures its products in: the stored setting
// when it names a unit of that organization, else the organization's own default unit by symbol
// (kg, cm). The fallback covers an organization whose settings were never written and a caller the
// settings module cannot attribute to a single organization.
type OrgUnitResolver struct {
	settings itExt.EffectiveSettingsExtService
	uoms     itExt.UomDirectoryExtService
}

func NewOrgUnitResolver(
	settings itExt.EffectiveSettingsExtService, uoms itExt.UomDirectoryExtService,
) *OrgUnitResolver {
	return &OrgUnitResolver{settings: settings, uoms: uoms}
}

// ResolveUnits reads the units of several organizations at once: one settings read and one lookup
// per stored unit for all of them, then one symbol lookup per organization and missing unit.
func (this *OrgUnitResolver) ResolveUnits(ctx corectx.Context, orgIds []string) (map[string]OrgUnits, error) {
	stored, err := this.storedUnits(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]OrgUnits, len(orgIds))
	for _, orgId := range orgIds {
		units := OrgUnits{}
		for name, symbol := range invsettings.DefaultUomSymbolOf {
			if uom := stored[name]; uom != nil && string(uom.OrgId) == orgId {
				units[name] = uom
				continue
			}
			uom, err := this.findBySymbol(ctx, orgId, symbol)
			if err != nil {
				return nil, err
			}
			if uom != nil {
				units[name] = uom
			}
		}
		out[orgId] = units
	}
	return out, nil
}

// DefaultUnitIds resolves a new organization's default units to its own unit ids, for the org
// lifecycle to store. A default the organization has no unit for is left out.
func (this *OrgUnitResolver) DefaultUnitIds(ctx corectx.Context, orgId string) (map[string]string, error) {
	ids := map[string]string{}
	for name, symbol := range invsettings.DefaultUomSymbolOf {
		uom, err := this.findBySymbol(ctx, orgId, symbol)
		if err != nil {
			return nil, err
		}
		if uom != nil {
			ids[name] = string(uom.Id)
		}
	}
	return ids, nil
}

// AssertUnitSetting checks a new value of a unit setting: the unit exists, is not archived, and is
// of the same category and organization as the unit the setting stands for now (the stored one, or
// the organization's default). That keeps a weight setting on a weight unit without the UoM
// categories carrying a kind: the first value is the kg the org lifecycle wrote.
func (this *OrgUnitResolver) AssertUnitSetting(
	ctx corectx.Context, name string, uomId string,
) (*ft.ClientErrorItem, error) {
	uom, err := this.findById(ctx, uomId)
	if err != nil {
		return nil, err
	}
	if uom == nil {
		return ft.NewValidationError(name, ft.ErrorKey("inventory_settings.uom_not_found", "inventory"),
			"the unit of measure does not exist"), nil
	}
	if uom.IsArchived {
		return ft.NewValidationError(name, ft.ErrorKey("inventory_settings.uom_archived", "inventory"),
			"the unit of measure is archived"), nil
	}
	reference, err := this.referenceUnit(ctx, name, string(uom.OrgId))
	if err != nil || reference == nil {
		return nil, err
	}
	if reference.CategoryId != uom.CategoryId || reference.OrgId != uom.OrgId {
		return ft.NewValidationError(name, ft.ErrorKey("inventory_settings.uom_wrong_category", "inventory"),
			"choose a unit of the same kind and organization as {{current}}",
			map[string]any{"current": reference.Symbol}), nil
	}
	return nil, nil
}

func (this *OrgUnitResolver) referenceUnit(
	ctx corectx.Context, name string, orgId string,
) (*itExt.GetUomResultData, error) {
	stored, err := this.storedUnits(ctx)
	if err != nil {
		return nil, err
	}
	if uom := stored[name]; uom != nil {
		return uom, nil
	}
	return this.findBySymbol(ctx, orgId, invsettings.DefaultUomSymbolOf[name])
}

// storedUnits reads the acting organization's unit settings and the units they name.
func (this *OrgUnitResolver) storedUnits(ctx corectx.Context) (OrgUnits, error) {
	result, err := this.settings.GetEffectiveSettings(ctx, itExt.GetEffectiveSettingsQuery{
		ModuleKeys: []string{invconstants.InventoryModuleName},
	})
	if err != nil {
		return nil, errors.Wrap(err, "read inventory unit settings")
	}
	units := OrgUnits{}
	if result == nil || !result.HasData || result.ClientErrors.Count() > 0 {
		return units, nil
	}
	for name := range invsettings.DefaultUomSymbolOf {
		uomId, _ := result.Data.Values[invconstants.InventoryModuleName+"."+name].(string)
		if uomId == "" {
			continue
		}
		uom, err := this.findById(ctx, uomId)
		if err != nil {
			return nil, err
		}
		if uom != nil {
			units[name] = uom
		}
	}
	return units, nil
}

func (this *OrgUnitResolver) findById(ctx corectx.Context, uomId string) (*itExt.GetUomResultData, error) {
	found, err := this.uoms.GetUom(ctx, itExt.GetUomQuery{Id: model.Id(uomId)})
	if err != nil {
		return nil, errors.Wrap(err, "read unit of measure")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return &found.Data, nil
}

func (this *OrgUnitResolver) findBySymbol(
	ctx corectx.Context, orgId string, symbol string,
) (*itExt.GetUomResultData, error) {
	if orgId == "" || symbol == "" {
		return nil, nil
	}
	found, err := this.uoms.FindUomBySymbol(ctx, itExt.FindUomBySymbolQuery{OrgId: model.Id(orgId), Symbol: symbol})
	if err != nil {
		return nil, errors.Wrap(err, "find unit of measure by symbol")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return &found.Data, nil
}
