package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

// lockOwned locks the definition and reads it, answering not-found for a definition of another org:
// lookups outside the caller's ownership disclose nothing (BR §9.4).
func (this *SequenceDomainServiceImpl) lockOwned(
	ctx corectx.Context, id model.Id, orgId *model.Id, mode itSeq.LockMode,
) (dmodel.DynamicFields, *ft.ClientErrors, error) {
	found, err := this.definitionRepo.LockDefinition(ctx, id, mode)
	if err != nil {
		return nil, nil, err
	}
	var stored dmodel.DynamicFields
	if found {
		if stored, err = fetchDefinition(ctx, this.definitionRepo, id); err != nil {
			return nil, nil, err
		}
	}
	if stored == nil || !ownedBy(stored, orgId) {
		cErrs := singleError(ft.NewNotFoundError(basemodel.FieldId))
		return nil, &cErrs, nil
	}
	return stored, nil, nil
}

// ownedBy: an org definition serves only its org; a tenant definition serves every org.
func ownedBy(stored dmodel.DynamicFields, orgId *model.Id) bool {
	defOrg := models.FieldString(stored, basemodel.FieldOrgId)
	if defOrg == "" {
		return true
	}
	return orgId != nil && string(*orgId) == defOrg
}
