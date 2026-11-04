package sequence

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// DefinitionRegistrar is the part of a consumer's sequence port that registers definitions.
type DefinitionRegistrar interface {
	UpsertDefinition(ctx corectx.Context, cmd UpsertDefinitionCommand) (*UpsertDefinitionResult, error)
}

// OrgSettingsSeeder is the part of a consumer's settings port that writes its own org settings
// from the org lifecycle.
type OrgSettingsSeeder interface {
	SeedOrgSettings(ctx corectx.Context, cmd itSettings.SetSettingsCommand) (*itSettings.SetSettingsResult, error)
}

// OrgRegistration is one definition a new organization gets, and the consumer's org setting its id
// is recorded in.
type OrgRegistration struct {
	Setting string
	Command UpsertDefinitionCommand
}

// RegisterOrgDefinitions registers a consumer's definitions for one organization and records their
// ids in its org settings under moduleKey, returning the ids keyed by setting. Only the org
// lifecycle calls it: a definition is never registered while a record is processed. Upsert returns
// an existing definition unchanged, so a re-run lifecycle points the settings at the same series.
// A refusal is the consumer's own defect (it built the command), so it is an error.
func RegisterOrgDefinitions(
	ctx corectx.Context, registrar DefinitionRegistrar, seeder OrgSettingsSeeder,
	moduleKey string, orgId model.Id, registrations ...OrgRegistration,
) (map[string]model.Id, error) {
	ids := make(map[string]model.Id, len(registrations))
	items := make([]itSettings.SetSettingItem, 0, len(registrations))
	for _, registration := range registrations {
		definitionId, err := upsertDefinition(ctx, registrar, registration.Command)
		if err != nil {
			return nil, err
		}
		ids[registration.Setting] = definitionId
		items = append(items, itSettings.SetSettingItem{Name: registration.Setting, Value: string(definitionId)})
	}
	seeded, err := seeder.SeedOrgSettings(ctx, itSettings.SetSettingsCommand{
		ModuleKey: moduleKey, OrgId: &orgId, Items: items,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "record the %s sequences", moduleKey)
	}
	if seeded.ClientErrors.Count() > 0 {
		return nil, errors.Wrapf(seeded.ClientErrors.ToError(), "record the %s sequences", moduleKey)
	}
	return ids, nil
}

func upsertDefinition(
	ctx corectx.Context, registrar DefinitionRegistrar, cmd UpsertDefinitionCommand,
) (model.Id, error) {
	result, err := registrar.UpsertDefinition(ctx, cmd)
	if err != nil {
		return "", errors.Wrapf(err, "register sequence %s", cmd.Code)
	}
	if result.ClientErrors.Count() > 0 {
		return "", errors.Wrapf(result.ClientErrors.ToError(), "register sequence %s", cmd.Code)
	}
	if !result.HasData || result.Data.DefinitionId == "" {
		return "", errors.Errorf("register sequence %s returned no definition", cmd.Code)
	}
	return result.Data.DefinitionId, nil
}

// ErrKeyNotConfigured is the translation key, in the consumer's own namespace, of a document refused
// because its organization holds no definition id for it: an org whose lifecycle never ran, or one
// created before the definition existed and missed by the seeds.
const ErrKeyNotConfigured = "sequence_not_configured"

// NotConfiguredError is the client error a consumer returns, under its own module name, when the
// org setting naming a definition is unset. A consumer never registers the definition there.
func NotConfiguredError(moduleName string, setting string) *ft.ClientErrorItem {
	return ft.NewAnonymousBusinessViolation(ft.ErrorKey(ErrKeyNotConfigured, moduleName),
		"document numbering is not set up for this organization ({{setting}} is unset)",
		map[string]any{"setting": setting})
}
