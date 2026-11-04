package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

func NewSequenceDomainServiceImpl(
	definitionSvc itSeq.SequenceDefinitionDomainService,
	counterRepo itSeq.SequenceCounterRepository,
	issuanceRepo itSeq.SequenceIssuanceRepository,
) itSeq.SequenceDomainService {
	return &SequenceDomainServiceImpl{
		definitionSvc:  definitionSvc,
		definitionRepo: definitionSvc.Repository().(itSeq.SequenceDefinitionRepository),
		counterRepo:    counterRepo,
		issuanceRepo:   issuanceRepo,
	}
}

// SequenceDomainServiceImpl is what the port runs: registering a definition, issuing a number and
// releasing a definition. Every operation joins the caller's transaction, or opens one.
type SequenceDomainServiceImpl struct {
	definitionSvc  itSeq.SequenceDefinitionDomainService
	definitionRepo itSeq.SequenceDefinitionRepository
	counterRepo    itSeq.SequenceCounterRepository
	issuanceRepo   itSeq.SequenceIssuanceRepository
}

// UpsertDefinition returns the definition registered under (configuration_scope, org, code), or
// creates it (SEQ-DEF-03/04). Concurrent registrations of one key serialize on an advisory lock and
// so resolve to the same id; an existing definition is never changed, unarchived or reset.
func (this *SequenceDomainServiceImpl) UpsertDefinition(
	ctx corectx.Context, cmd itSeq.UpsertDefinitionCommand,
) (*itSeq.UpsertDefinitionResult, error) {
	var data itSeq.UpsertDefinitionResultData
	cErrs, err := dynamicresource.InRefusableTransaction(ctx, this.definitionRepo,
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var cErrs *ft.ClientErrors
			var err error
			data, cErrs, err = this.upsert(tranxCtx, cmd)
			return cErrs, err
		})
	if err != nil {
		return nil, err
	}
	if cErrs != nil {
		return &itSeq.UpsertDefinitionResult{ClientErrors: *cErrs}, nil
	}
	return &itSeq.UpsertDefinitionResult{Data: data, HasData: true}, nil
}

func (this *SequenceDomainServiceImpl) upsert(
	ctx corectx.Context, cmd itSeq.UpsertDefinitionCommand,
) (itSeq.UpsertDefinitionResultData, *ft.ClientErrors, error) {
	orgId := upsertOrgId(cmd)
	key := cmd.ConfigurationScope + "|" + derefString(orgId) + "|" + cmd.Code
	if err := this.definitionRepo.LockRegistration(ctx, key); err != nil {
		return itSeq.UpsertDefinitionResultData{}, nil, err
	}
	existingId, err := this.definitionRepo.FindIdByCode(ctx, cmd.Code, orgId)
	if err != nil {
		return itSeq.UpsertDefinitionResultData{}, nil, err
	}
	if existingId != "" {
		cErrs, err := this.assertSameContract(ctx, existingId, cmd)
		return itSeq.UpsertDefinitionResultData{DefinitionId: existingId}, cErrs, err
	}
	created, err := this.definitionSvc.Create(ctx, UpsertFields(cmd))
	if err != nil {
		return itSeq.UpsertDefinitionResultData{}, nil, err
	}
	if created.ClientErrors.Count() > 0 {
		return itSeq.UpsertDefinitionResultData{}, &created.ClientErrors, nil
	}
	createdId := models.FieldString(created.Data, basemodel.FieldId)
	return itSeq.UpsertDefinitionResultData{DefinitionId: model.Id(createdId), Created: true}, nil, nil
}

// assertSameContract refuses a registration whose owner or scope contradicts the stored one.
func (this *SequenceDomainServiceImpl) assertSameContract(
	ctx corectx.Context, id model.Id, cmd itSeq.UpsertDefinitionCommand,
) (*ft.ClientErrors, error) {
	stored, err := fetchDefinition(ctx, this.definitionRepo, id)
	if err != nil || stored == nil {
		return nil, err
	}
	def := models.NewSequenceDefinitionFrom(stored)
	wantScope := ScopeTypeOf(cmd.ConfigurationScope, derefString(cmd.SubScope))
	if def.GetOwnerModule() == cmd.OwnerModule && def.GetScopeType() == wantScope {
		return nil, nil
	}
	cErrs := singleError(newSeqViolation(models.SeqDefFieldCode, ErrKeySeqRegistrationConflict,
		"'{{code}}' is already registered by {{owner}} with scope {{scope}}",
		map[string]any{"code": cmd.Code, "owner": def.GetOwnerModule(), "scope": def.GetScopeType()}))
	return &cErrs, nil
}

// UpsertFields maps a registration onto the definition's create command.
func UpsertFields(cmd itSeq.UpsertDefinitionCommand) dmodel.DynamicFields {
	fields := dmodel.DynamicFields{
		models.SeqDefFieldConfigurationScope: cmd.ConfigurationScope,
		models.SeqDefFieldOwnerModule:        cmd.OwnerModule,
		models.SeqDefFieldCode:               cmd.Code,
		models.SeqDefFieldName:               cmd.Name,
	}
	if orgId := upsertOrgId(cmd); orgId != nil {
		fields[models.SeqDefFieldOrgId] = string(*orgId)
	}
	putIfSet(fields, models.SeqDefFieldDescription, cmd.Description)
	putIfSet(fields, models.SeqDefFieldSubScope, cmd.SubScope)
	putIfSet(fields, models.SeqDefFieldFormatOwner, cmd.FormatOwner)
	putIfSet(fields, models.SeqDefFieldAlgorithm, cmd.Algorithm)
	putIfSet(fields, models.SeqDefFieldPrefixTemplate, cmd.PrefixTemplate)
	putIfSet(fields, models.SeqDefFieldSuffixTemplate, cmd.SuffixTemplate)
	putIfSet(fields, models.SeqDefFieldPadding, cmd.Padding)
	putIfSet(fields, models.SeqDefFieldIncrement, cmd.Increment)
	putIfSet(fields, models.SeqDefFieldInitialValue, cmd.InitialValue)
	putIfSet(fields, models.SeqDefFieldResetPeriod, cmd.ResetPeriod)
	putIfSet(fields, models.SeqDefFieldTimezone, cmd.Timezone)
	putIfSet(fields, models.SeqDefFieldIssuanceRetentionDays, cmd.IssuanceRetentionDays)
	return fields
}

func putIfSet[T any](fields dmodel.DynamicFields, key string, value *T) {
	if value != nil {
		fields[key] = *value
	}
}

// upsertOrgId is the org a registration belongs to: none for a tenant-wide definition.
func upsertOrgId(cmd itSeq.UpsertDefinitionCommand) *model.Id {
	if cmd.ConfigurationScope == models.SeqConfigScopeTenant || cmd.OrgId == nil || *cmd.OrgId == "" {
		return nil
	}
	return cmd.OrgId
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
