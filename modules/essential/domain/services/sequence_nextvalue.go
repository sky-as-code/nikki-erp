package services

import (
	"strings"
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	seqalgo "github.com/sky-as-code/nikki-erp/modules/essential/domain/services/sequence"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

const seqMaxOperationKeyLength = itSeq.MaxOperationKeyLength

// nextValueRequest is a NextValue command after normalization, outside any lock (BR §6.3 step 1).
type nextValueRequest struct {
	cmd           itSeq.NextValueCommand
	idempotencyId string
	operationKey  string
	effectiveAt   *time.Time
	acceptedAt    time.Time
	tenantId      string
	principalId   string
}

// NextValue issues one number (BR §6.3). The request key is reserved before anything else, so a
// retry finds the committed issuance and replays it until its expires_at (after which it answers
// idempotency_expired, and a different fingerprint answers idempotency_conflict); a new request
// reads the definition under a share lock, moves its counter with one atomic upsert and completes
// the issuance in the same transaction.
func (this *SequenceDomainServiceImpl) NextValue(
	ctx corectx.Context, cmd itSeq.NextValueCommand,
) (*itSeq.NextValueResult, error) {
	req, cErrs := newNextValueRequest(ctx, cmd)
	if cErrs != nil {
		return &itSeq.NextValueResult{ClientErrors: *cErrs}, nil
	}
	var data itSeq.NextValueResultData
	cErrs, err := dynamicresource.InRefusableTransaction(ctx, this.definitionRepo,
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var cErrs *ft.ClientErrors
			var err error
			data, cErrs, err = this.issue(tranxCtx, req)
			return cErrs, err
		})
	if err != nil {
		return nil, err
	}
	if cErrs != nil {
		return &itSeq.NextValueResult{ClientErrors: *cErrs}, nil
	}
	return &itSeq.NextValueResult{Data: data, HasData: true}, nil
}

func newNextValueRequest(ctx corectx.Context, cmd itSeq.NextValueCommand) (nextValueRequest, *ft.ClientErrors) {
	req := nextValueRequest{
		cmd: cmd, acceptedAt: time.Now().UTC(), tenantId: actingTenantId(ctx), principalId: actingPrincipalId(ctx),
	}
	vErrs := ft.NewClientErrors()
	if cmd.DefinitionId == "" {
		vErrs.Append(*newSeqValidation("definition_id", ErrKeySeqInvalidInput, "definition_id is required"))
	}
	id, err := corectx.ParseIdempotencyId(derefString(cmd.IdempotencyId))
	if err != nil {
		vErrs.Append(*newSeqValidation("idempotency_id", ErrKeySeqInvalidIdempotencyId,
			"the idempotency id must be a canonical UUIDv7"))
	}
	req.idempotencyId = id
	req.operationKey = strings.TrimSpace(derefString(cmd.OperationKey))
	if req.operationKey == "" {
		req.operationKey = models.SeqDefaultOperationKey
	}
	if len(req.operationKey) > seqMaxOperationKeyLength {
		vErrs.Append(*newSeqValidation("operation_key", ErrKeySeqInvalidInput,
			"operation_key is at most {{max}} characters", map[string]any{"max": seqMaxOperationKeyLength}))
	}
	if cmd.EffectiveAt != nil && *cmd.EffectiveAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, *cmd.EffectiveAt)
		if err != nil {
			vErrs.Append(*newSeqValidation("effective_at", ErrKeySeqInvalidInput,
				"effective_at must be an RFC 3339 timestamp with a time zone"))
		}
		req.effectiveAt = &parsed
	}
	if vErrs.Count() > 0 {
		return req, vErrs
	}
	return req, nil
}

func (this *SequenceDomainServiceImpl) issue(
	ctx corectx.Context, req nextValueRequest,
) (itSeq.NextValueResultData, *ft.ClientErrors, error) {
	none := itSeq.NextValueResultData{}
	stored, err := fetchDefinition(ctx, this.definitionRepo, req.cmd.DefinitionId)
	if err != nil {
		return none, nil, err
	}
	if stored == nil || !ownedBy(stored, req.cmd.OrgId) {
		cErrs := singleError(ft.NewNotFoundError("definition_id"))
		return none, &cErrs, nil
	}
	def := models.NewSequenceDefinitionFrom(stored)
	issuanceId, err := model.NewId()
	if err != nil {
		return none, nil, err
	}
	fingerprint := requestFingerprint(def, req)
	reserved, err := this.issuanceRepo.ReserveIssuance(ctx, reserveParam(*issuanceId, def, req, fingerprint))
	if err != nil {
		return none, nil, err
	}
	if !reserved {
		return this.replay(ctx, req, fingerprint)
	}
	return this.issueNew(ctx, req, *issuanceId, def)
}

func requestFingerprint(def *models.SequenceDefinition, req nextValueRequest) string {
	return seqalgo.Fingerprint(seqalgo.FingerprintInput{
		DefinitionId:  string(def.GetId()),
		TenantId:      req.tenantId,
		OrgId:         def.GetOrgId(),
		OwnerModule:   def.GetOwnerModule(),
		CallerModule:  req.cmd.CallerModule,
		ScopeObjectId: derefString(req.cmd.ScopeObjectId),
		OperationKey:  req.operationKey,
		EffectiveAt:   req.effectiveAt,
	})
}

func reserveParam(
	id model.Id, def *models.SequenceDefinition, req nextValueRequest, fingerprint string,
) itSeq.ReserveIssuanceParam {
	return itSeq.ReserveIssuanceParam{
		NewId:              id,
		DefinitionId:       def.GetId(),
		OrgId:              optionalId(def.GetOrgId()),
		IdempotencyId:      req.idempotencyId,
		OperationKey:       req.operationKey,
		RequestFingerprint: fingerprint,
		CallerModule:       req.cmd.CallerModule,
		CallerPrincipalId:  optionalId(req.principalId),
		AcceptedAt:         req.acceptedAt,
	}
}
