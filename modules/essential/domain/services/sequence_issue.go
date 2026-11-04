package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	seqalgo "github.com/sky-as-code/nikki-erp/modules/essential/domain/services/sequence"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

// issueNew runs BR §6.3 steps 4-8 for a request whose key this transaction just reserved.
func (this *SequenceDomainServiceImpl) issueNew(
	ctx corectx.Context, req nextValueRequest, issuanceId model.Id, preview *models.SequenceDefinition,
) (itSeq.NextValueResultData, *ft.ClientErrors, error) {
	none := itSeq.NextValueResultData{}
	if cErrs := assertIssuable(preview); cErrs != nil {
		return none, cErrs, nil
	}
	def, cErrs, err := this.lockForIssue(ctx, req)
	if err != nil || cErrs != nil {
		return none, cErrs, err
	}
	scopeKey, cErrs := BuildScopeKey(def, req.tenantId, req.cmd.ScopeObjectId)
	if cErrs != nil {
		return none, cErrs, nil
	}
	effectiveAt := req.acceptedAt
	if req.effectiveAt != nil {
		effectiveAt = req.effectiveAt.UTC()
	}
	period, err := seqalgo.PeriodOf(effectiveAt, def.GetTimezone(), def.GetResetPeriod())
	if err != nil {
		return none, nil, errors.Wrap(err, "NextValue.PeriodOf")
	}
	counter, cErrs, err := this.increment(ctx, def, req, scopeKey, period.Key)
	if err != nil || cErrs != nil {
		return none, cErrs, err
	}
	formatted, cErrs, err := FormatNumber(def, counter, seqalgo.TemplateVars{
		EffectiveAt: effectiveAt, AcceptedAt: req.acceptedAt, RangeStart: period.Start, PeriodKey: period.Key,
	})
	if err != nil || cErrs != nil {
		return none, cErrs, err
	}
	issuedAt := time.Now().UTC()
	complete := itSeq.CompleteIssuanceParam{
		Id: issuanceId, CounterId: counter.Id, EffectiveAt: effectiveAt, AllocatedValue: counter.LastNumber,
		FormattedValue: formatted, IssuedAt: issuedAt,
		ExpiresAt: issuedAt.Add(time.Duration(def.GetRetentionDays()) * 24 * time.Hour),
	}
	if err := this.issuanceRepo.CompleteIssuance(ctx, complete); err != nil {
		return none, nil, err
	}
	return newResultData(def, req, complete, scopeKey, period.Key), nil, nil
}

// assertIssuable refuses an archived definition before anything is locked or stamped.
func assertIssuable(def *models.SequenceDefinition) *ft.ClientErrors {
	if !def.IsArchived() {
		return nil
	}
	cErrs := singleError(newSeqViolation("definition_id", ErrKeySeqArchived,
		"the sequence is archived and issues no new numbers"))
	return &cErrs
}

// lockForIssue locks the definition and rereads it. The first issuance stamps first_issued_at, an
// UPDATE that is also its exclusive lock; later ones take FOR SHARE and run side by side. Stamping
// before any share lock is held is what keeps two racing first issuances from deadlocking.
func (this *SequenceDomainServiceImpl) lockForIssue(
	ctx corectx.Context, req nextValueRequest,
) (*models.SequenceDefinition, *ft.ClientErrors, error) {
	stamped, err := this.definitionRepo.StampFirstIssued(ctx, req.cmd.DefinitionId, req.acceptedAt)
	if err != nil {
		return nil, nil, err
	}
	mode := itSeq.LockModeShare
	if stamped {
		mode = itSeq.LockModeUpdate
	}
	stored, cErrs, err := this.lockOwned(ctx, req.cmd.DefinitionId, req.cmd.OrgId, mode)
	if err != nil || cErrs != nil {
		return nil, cErrs, err
	}
	def := models.NewSequenceDefinitionFrom(stored)
	return def, assertIssuable(def), nil
}

func (this *SequenceDomainServiceImpl) increment(
	ctx corectx.Context, def *models.SequenceDefinition, req nextValueRequest, scopeKey string, periodKey string,
) (*itSeq.IncrementedCounter, *ft.ClientErrors, error) {
	counterId, err := model.NewId()
	if err != nil {
		return nil, nil, err
	}
	param := itSeq.IncrementCounterParam{
		NewId: *counterId, DefinitionId: def.GetId(), OrgId: optionalId(def.GetOrgId()),
		ScopeObjectId: req.cmd.ScopeObjectId, ScopeKey: scopeKey, PeriodKey: periodKey,
		InitialValue: def.GetInitialValue(), Increment: def.GetIncrement(),
	}
	if def.GetAlgorithm() == models.SeqAlgorithmFf1Fpe {
		// Generated every call, kept only by the insert that creates the counter.
		tweak, err := seqalgo.NewTweakHex()
		if err != nil {
			return nil, nil, err
		}
		param.AlgorithmParams = dmodel.DynamicFields{models.SeqAlgorithmParamTweak: tweak}
	}
	counter, err := this.counterRepo.IncrementCounter(ctx, param)
	if err != nil {
		return nil, nil, err
	}
	if counter == nil {
		cErrs := singleError(newSeqViolation("definition_id", ErrKeySeqExhausted,
			"the counter has no room for another increment"))
		return nil, &cErrs, nil
	}
	return counter, nil, nil
}

// replay answers a request whose key already holds a committed issuance (SEQ-IDEM-03, §6.4).
func (this *SequenceDomainServiceImpl) replay(
	ctx corectx.Context, req nextValueRequest, fingerprint string,
) (itSeq.NextValueResultData, *ft.ClientErrors, error) {
	none := itSeq.NextValueResultData{}
	stored, err := this.issuanceRepo.FindByKeys(ctx, dmodel.DynamicFields{
		models.SeqIssFieldDefinitionId:  string(req.cmd.DefinitionId),
		models.SeqIssFieldIdempotencyId: req.idempotencyId,
		models.SeqIssFieldOperationKey:  req.operationKey,
	})
	if err != nil {
		return none, nil, err
	}
	if stored == nil || !stored.HasData {
		return none, nil, errors.New("NextValue: the request key is taken but its issuance cannot be read")
	}
	if cErrs := assertReplayable(stored.Data, fingerprint); cErrs != nil {
		return none, cErrs, nil
	}
	counterId := models.FieldString(stored.Data, models.SeqIssFieldCounterId)
	counter, err := this.counterRepo.FindByKeys(ctx, dmodel.DynamicFields{basemodel.FieldId: counterId})
	if err != nil {
		return none, nil, err
	}
	if counter == nil || !counter.HasData {
		return none, nil, errors.Errorf("NextValue: the counter '%s' of a stored issuance is missing", counterId)
	}
	return replayedData(stored.Data, counter.Data, req), nil, nil
}

func assertReplayable(issuance dmodel.DynamicFields, fingerprint string) *ft.ClientErrors {
	if models.FieldString(issuance, models.SeqIssFieldRequestFingerprint) != fingerprint {
		cErrs := singleError(newSeqViolation("idempotency_id", ErrKeySeqIdempotencyConflict,
			"this idempotency id was already used for a different request"))
		return &cErrs
	}
	expiresAt := issuance.GetModelDateTime(models.SeqIssFieldExpiresAt)
	if expiresAt == nil || !time.Now().Before(expiresAt.GoTime()) {
		cErrs := singleError(newSeqViolation("idempotency_id", ErrKeySeqIdempotencyExpired,
			"the issuance of this request has expired; reconcile the business record instead of retrying"))
		return &cErrs
	}
	return nil
}
