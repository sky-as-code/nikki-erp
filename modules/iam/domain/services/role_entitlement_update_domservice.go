package services

import (
	"fmt"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

// validateEntitlementUpdates refuses, before anything is written, an update without an id, two
// updates of the same entitlement, and an update of an entitlement the same delta removes.
func validateEntitlementUpdates(delta itRole.RoleEntitlementsDelta) ft.ClientErrors {
	removed := make(map[model.Id]bool, len(delta.Remove))
	for _, entId := range delta.Remove {
		removed[entId] = true
	}
	seen := make(map[model.Id]bool, len(delta.Update))
	var out ft.ClientErrors
	for i, update := range delta.Update {
		field := fmt.Sprintf("%s[%d]", deltaFieldUpdate, i)
		switch {
		case update.Id == "":
			out = append(out, *dmodel.NewMissingFieldErr(field+"."+basemodel.FieldId))
		case removed[update.Id]:
			out = append(out, *ft.NewBusinessViolation(field, ft.ErrorKey("err_entitlement_update_removed", "iam"),
				"entitlement {{id}} is removed by the same request", map[string]any{"id": string(update.Id)}))
		case seen[update.Id]:
			out = append(out, *ft.NewBusinessViolation(field, ft.ErrorKey("err_entitlement_update_repeated", "iam"),
				"entitlement {{id}} is updated twice", map[string]any{"id": string(update.Id)}))
		}
		seen[update.Id] = true
	}
	return out
}

func (this *RoleDomainServiceImpl) updateRoleEntitlements(
	ctx corectx.Context, roleId model.Id, updates []itRole.RoleEntitlementUpdate,
) (ft.ClientErrors, error) {
	if len(updates) == 0 {
		return nil, nil
	}
	etags, cErrs, err := this.roleEntitlementEtags(ctx, roleId, updates)
	if err != nil || cErrs.Count() > 0 {
		return prefixFields(cErrs, deltaFieldUpdate), err
	}
	if out := assertUpdatesOwned(updates, etags); out.Count() > 0 {
		return out, nil
	}
	for i, update := range updates {
		res, err := this.entitlementSvc.UpdateEntitlementDeferRebuild(ctx, entitlementUpdateCommand(update, etags[update.Id]))
		if err != nil {
			return nil, err
		}
		if res.ClientErrors.Count() > 0 {
			return prefixFields(res.ClientErrors, fmt.Sprintf("%s[%d]", deltaFieldUpdate, i)), nil
		}
	}
	return nil, nil
}

func assertUpdatesOwned(updates []itRole.RoleEntitlementUpdate, etags map[model.Id]model.Etag) ft.ClientErrors {
	var out ft.ClientErrors
	for i, update := range updates {
		if _, isOwned := etags[update.Id]; !isOwned {
			out = append(out, *ft.NewBusinessViolation(
				fmt.Sprintf("%s[%d]", deltaFieldUpdate, i), ft.ErrorKey("err_entitlement_not_in_role", "iam"),
				"entitlement {{id}} does not belong to this role", map[string]any{"id": string(update.Id)},
			))
		}
	}
	return out
}

// roleEntitlementEtags reads the current etag of every entitlement to update, so the update can
// go through the entitlement service's optimistic check without the client sending etags.
func (this *RoleDomainServiceImpl) roleEntitlementEtags(
	ctx corectx.Context, roleId model.Id, updates []itRole.RoleEntitlementUpdate,
) (map[model.Id]model.Etag, ft.ClientErrors, error) {
	ids := make([]any, len(updates))
	for i := range updates {
		ids[i] = updates[i].Id
	}
	graph := dmodel.NewSearchGraph().And(
		*dmodel.NewSearchNode().Condition(dmodel.NewCondition(domain.EntitlementFieldRoleId, dmodel.Equals, roleId)),
		*dmodel.NewSearchNode().Condition(dmodel.NewCondition(basemodel.FieldId, dmodel.In, ids...)),
	)
	searchRes, err := this.entitlementRepo.Search(ctx, dyn.RepoSearchParam{
		Graph:  graph,
		Fields: []string{basemodel.FieldId, basemodel.FieldEtag},
		Page:   0,
		Size:   len(ids),
	})
	if err != nil {
		return nil, nil, err
	}
	if searchRes.ClientErrors.Count() > 0 {
		return nil, searchRes.ClientErrors, nil
	}
	etags := make(map[model.Id]model.Etag, len(searchRes.Data.Items))
	for _, ent := range searchRes.Data.Items {
		fields := ent.GetFieldData()
		// fmt.Sprint covers both shapes a value may come back in: string and its named type.
		etags[model.Id(fmt.Sprint(fields[basemodel.FieldId]))] = model.Etag(fmt.Sprint(fields[basemodel.FieldEtag]))
	}
	return etags, nil, nil
}

func entitlementUpdateCommand(update itRole.RoleEntitlementUpdate, etag model.Etag) itEnt.UpdateEntitlementCommand {
	ent := domain.NewEntitlement()
	entId := update.Id
	ent.SetId(&entId)
	ent.SetEtag(&etag)
	ent.GetFieldData().SetString(domain.EntitlementFieldDescription, update.Description)
	return itEnt.UpdateEntitlementCommand{Entitlement: *ent}
}
