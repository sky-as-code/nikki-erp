package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	modconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

// SequenceDefinitionDeps are what the definition rules need beyond the composable default.
type SequenceDefinitionDeps struct {
	UsageDispatcher *usagecheck.Dispatcher
	DefaultTimezone SequenceTimezoneDefaulter
}

func NewSequenceDefinitionDomainService(
	base dynamicresource.CrudDomainService, deps SequenceDefinitionDeps,
) itSeq.SequenceDefinitionDomainService {
	return &SequenceDefinitionDomainServiceImpl{CrudDomainService: base, deps: deps}
}

// SequenceDefinitionDomainServiceImpl attaches the definition rules to the built-in writes. Update
// and archive take the row FOR UPDATE first, so they serialize against an issuance, which holds the
// row FOR SHARE (SEQ-CFG-02, BR §4.4): an update that loses the race to the first issuance then sees
// first_issued_at and is refused.
type SequenceDefinitionDomainServiceImpl struct {
	dynamicresource.CrudDomainService

	deps SequenceDefinitionDeps
}

func (this *SequenceDefinitionDomainServiceImpl) definitionRepo() itSeq.SequenceDefinitionRepository {
	return this.Repository().(itSeq.SequenceDefinitionRepository)
}

func (this *SequenceDefinitionDomainServiceImpl) Create(
	ctx corectx.Context, cmd dynamicresource.CreateCommand, options ...dynamicresource.CreateOptions,
) (*dynamicresource.CreateResult, error) {
	opts := safe.GetOptional(options, dynamicresource.CreateOptions{})
	opts.BeforeValidation = dynamicresource.ChainBeforeValidation(normalizeDefinitionCreate(this.deps.DefaultTimezone),
		opts.BeforeValidation)
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(opts.ValidateExtra, validateDefinitionCreate())
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *SequenceDefinitionDomainServiceImpl) Update(
	ctx corectx.Context, cmd dynamicresource.UpdateCommand, options ...dynamicresource.UpdateOptions,
) (*dynamicresource.MutateResult, error) {
	opts := safe.GetOptional(options, dynamicresource.UpdateOptions{})
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(opts.ValidateExtra,
		validateDefinitionUpdate(this.Schema()))
	var result *dynamicresource.MutateResult
	err := dynamicresource.InTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		if found, err := this.definitionRepo().LockDefinition(
			tranxCtx, dynamicresource.ParamId(cmd), itSeq.LockModeUpdate); err != nil || !found {
			result = notFoundMutation()
			return err
		}
		var err error
		result, err = this.CrudDomainService.Update(tranxCtx, cmd, opts)
		return err
	})
	return result, err
}

// SetArchived asks every Essential dependant whether the definition is still configured for use
// before archiving it (BR §4.4); an unanswered check refuses. Unarchiving needs no check and keeps
// every counter.
func (this *SequenceDefinitionDomainServiceImpl) SetArchived(
	ctx corectx.Context, cmd dynamicresource.SetArchivedCommand,
) (*dynamicresource.MutateResult, error) {
	var result *dynamicresource.MutateResult
	err := dynamicresource.InTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		id := dynamicresource.ParamId(cmd)
		if found, err := this.definitionRepo().LockDefinition(tranxCtx, id, itSeq.LockModeUpdate); err != nil || !found {
			result = notFoundMutation()
			return err
		}
		if archived, _ := dynamicresource.ParamBool(cmd, basemodel.FieldIsArchived); archived {
			vErrs := ft.NewClientErrors()
			if err := this.assertNotInUse(tranxCtx, id, vErrs); err != nil || vErrs.Count() > 0 {
				result = &dynamicresource.MutateResult{ClientErrors: *vErrs}
				return err
			}
		}
		var err error
		result, err = this.CrudDomainService.SetArchived(tranxCtx, cmd)
		return err
	})
	return result, err
}

func (this *SequenceDefinitionDomainServiceImpl) assertNotInUse(
	ctx corectx.Context, id model.Id, vErrs *ft.ClientErrors,
) error {
	stored, err := this.fetch(ctx, id)
	if err != nil || stored == nil {
		return err
	}
	return usagecheck.AssertDeletable(ctx, this.deps.UsageDispatcher, modconstants.EssentialModuleName,
		basemodel.FieldId, usagecheck.ResourceRef{
			ResourceName: usagecheck.ResourceSequenceDefinition,
			Identifier:   usagecheck.NewIdentifier(string(id), models.FieldString(stored, basemodel.FieldOrgId)),
		}, vErrs)
}

func (this *SequenceDefinitionDomainServiceImpl) fetch(ctx corectx.Context, id model.Id) (dmodel.DynamicFields, error) {
	return fetchDefinition(ctx, this.Repository(), id)
}

// fetchDefinition reads one definition by id, nil when it does not exist in the caller's tenant.
func fetchDefinition(
	ctx corectx.Context, repo dynamicresource.CrudRepository, id model.Id,
) (dmodel.DynamicFields, error) {
	found, err := repo.FindByKeys(ctx, dmodel.DynamicFields{basemodel.FieldId: string(id)})
	if err != nil {
		return nil, err
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data, nil
}

func notFoundMutation() *dynamicresource.MutateResult {
	return &dynamicresource.MutateResult{ClientErrors: singleError(ft.NewAnonymousNotFoundError())}
}
