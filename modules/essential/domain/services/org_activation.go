package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

const activationFieldIsActive = "is_active"

// maxActivationsPerOrg bounds one org's activation read; a catalogue far smaller than this is
// what a tenant-wide list of currencies or languages amounts to.
const maxActivationsPerOrg = 1000

// OrgActivation reads and writes one organization's activation of a tenant-wide catalogue entry
// (a currency, a language) through the junction's own domain service, so every change is audited.
// catalogueField names the junction's reference to the entry: currency_id, language_id.
type OrgActivation struct {
	junction       dynamicresource.CrudDomainService
	catalogueField string
}

func NewOrgActivation(junction dynamicresource.CrudDomainService, catalogueField string) *OrgActivation {
	return &OrgActivation{junction: junction, catalogueField: catalogueField}
}

// SetActive records whether orgId uses entryId, creating the row the first time.
func (this *OrgActivation) SetActive(ctx corectx.Context, orgId model.Id, entryId model.Id, active bool) error {
	found, err := this.junction.Repository().FindByKeys(ctx, this.keys(orgId, entryId))
	if err != nil {
		return errors.Wrap(err, "OrgActivation.SetActive")
	}
	if found == nil || !found.HasData {
		created, err := this.junction.Create(ctx, dmodel.DynamicFields{
			basemodel.FieldOrgId: string(orgId), this.catalogueField: string(entryId), activationFieldIsActive: active,
		})
		return mutationError(err, created)
	}
	if current, _ := found.Data[activationFieldIsActive].(bool); current == active {
		return nil
	}
	cmd := dmodel.DynamicFields{basemodel.FieldId: found.Data[basemodel.FieldId], activationFieldIsActive: active}
	if etag, ok := found.Data[basemodel.FieldEtag]; ok && etag != nil {
		cmd[basemodel.FieldEtag] = etag
	}
	updated, err := this.junction.Update(ctx, cmd)
	if err != nil {
		return errors.Wrap(err, "OrgActivation.SetActive")
	}
	if updated != nil && updated.ClientErrors.Count() > 0 {
		return errors.Wrap(updated.ClientErrors.ToError(), "OrgActivation.SetActive")
	}
	return nil
}

// ActiveIds lists the entries orgId has active.
func (this *OrgActivation) ActiveIds(ctx corectx.Context, orgId model.Id) ([]model.Id, error) {
	result, err := this.junction.Repository().Search(ctx, dyn.RepoSearchParam{
		Fields: []string{this.catalogueField},
		Filter: []dmodel.DynamicFields{{basemodel.FieldOrgId: string(orgId), activationFieldIsActive: true}},
		Size:   maxActivationsPerOrg,
	})
	if err != nil {
		return nil, errors.Wrap(err, "OrgActivation.ActiveIds")
	}
	if result.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(result.ClientErrors.ToError(), "OrgActivation.ActiveIds")
	}
	ids := make([]model.Id, 0, len(result.Data.Items))
	for _, row := range result.Data.Items {
		if id, ok := row[this.catalogueField].(string); ok && id != "" {
			ids = append(ids, model.Id(id))
		}
	}
	return ids, nil
}

// IsActive tells whether orgId has entryId active; no row reads as inactive.
func (this *OrgActivation) IsActive(ctx corectx.Context, orgId model.Id, entryId model.Id) (bool, error) {
	found, err := this.junction.Repository().FindByKeys(ctx, this.keys(orgId, entryId))
	if err != nil {
		return false, errors.Wrap(err, "OrgActivation.IsActive")
	}
	if found == nil || !found.HasData {
		return false, nil
	}
	active, _ := found.Data[activationFieldIsActive].(bool)
	return active, nil
}

// DeleteForOrg removes every activation row of orgId: the org's cleanup.
func (this *OrgActivation) DeleteForOrg(ctx corectx.Context, orgId model.Id) error {
	_, err := this.junction.DeleteByOrg(ctx, orgId)
	return errors.Wrap(err, "OrgActivation.DeleteForOrg")
}

func (this *OrgActivation) keys(orgId model.Id, entryId model.Id) dmodel.DynamicFields {
	return dmodel.DynamicFields{basemodel.FieldOrgId: string(orgId), this.catalogueField: string(entryId)}
}

func mutationError(err error, created *dynamicresource.CreateResult) error {
	if err != nil {
		return errors.Wrap(err, "OrgActivation.SetActive")
	}
	if created != nil && created.ClientErrors.Count() > 0 {
		return errors.Wrap(created.ClientErrors.ToError(), "OrgActivation.SetActive")
	}
	return nil
}

// CurrencyActivation and LanguageActivation tell the two junctions apart in the container.
type CurrencyActivation struct{ *OrgActivation }

type LanguageActivation struct{ *OrgActivation }

// Activations is what a catalogue's domain service needs of its junction; OrgActivation is it.
type Activations interface {
	SetActive(ctx corectx.Context, orgId model.Id, entryId model.Id, active bool) error
	ActiveIds(ctx corectx.Context, orgId model.Id) ([]model.Id, error)
	IsActive(ctx corectx.Context, orgId model.Id, entryId model.Id) (bool, error)
}

// activeChange is the `is_active` a write carried. The field is computed, so it is taken out of
// the command before schema validation and written to the acting org's junction row afterwards.
type activeChange struct {
	sent   bool
	active bool
}

func (this activeChange) deactivates() bool {
	return this.sent && !this.active
}

// popActiveChange removes is_active from cmd. A create that names none activates the new entry for
// the acting org (defaultOnCreate); an update that names none leaves the activation alone.
func popActiveChange(cmd dmodel.DynamicFields, defaultOnCreate bool) activeChange {
	raw, sent := cmd[activationFieldIsActive]
	delete(cmd, activationFieldIsActive)
	if !sent || raw == nil {
		return activeChange{sent: defaultOnCreate, active: defaultOnCreate}
	}
	active, _ := raw.(bool)
	return activeChange{sent: true, active: active}
}

// applyActiveChange writes the change for the acting org; without one there is no row to write.
func applyActiveChange(ctx corectx.Context, activation Activations, entryId any, change activeChange) error {
	orgId, ok := dyn.ActingOrgOf(ctx)
	id, _ := entryId.(string)
	if !change.sent || !ok || id == "" {
		return nil
	}
	return activation.SetActive(ctx, orgId, model.Id(id), change.active)
}

// resolveListingOrg is the acting org, else the caller's only org: a read with no org parameter
// (the user context) still answers for the organization a single-org user works in.
func resolveListingOrg(ctx corectx.Context) (model.Id, bool) {
	if orgId, ok := dyn.ActingOrgOf(ctx); ok {
		return orgId, true
	}
	orgIds := ctx.GetPermissions().UserOrgIds.ToSlice()
	if len(orgIds) == 1 {
		return orgIds[0], true
	}
	return "", false
}

func idsAsAny(ids []model.Id) []any {
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	return out
}
