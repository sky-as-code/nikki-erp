package services

import (
	"fmt"
	"strings"
	"time"

	"go.bryk.io/pkg/errors"
	"go.bryk.io/pkg/ulid"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	seqalgo "github.com/sky-as-code/nikki-erp/modules/essential/domain/services/sequence"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

const seqTenantKeyField = "tenant_id"

// BuildScopeKey is the canonical counter scope (BR §3.2): "tenant:<t>[/org:<o>][/<sub_scope>:<object>]".
// The tenant part is omitted in a binary without tenants. A custom scope requires the object id; a
// plain one refuses it, so a caller cannot split a counter the definition does not split.
func BuildScopeKey(
	def *models.SequenceDefinition, tenantId string, scopeObjectId *model.Id,
) (string, *ft.ClientErrors) {
	parts := make([]string, 0, 3)
	if tenantId != "" {
		parts = append(parts, "tenant:"+tenantId)
	}
	if def.GetConfigurationScope() == models.SeqConfigScopeOrg {
		parts = append(parts, "org:"+def.GetOrgId())
	}
	objectId := derefString(scopeObjectId)
	subScope := def.GetSubScope()
	if subScope == "" && objectId != "" {
		return "", invalidScopeObject("the definition has no custom scope; omit scope_object_id")
	}
	if subScope != "" {
		if !looksLikeUlid(objectId) {
			return "", invalidScopeObject("scope '{{scope}}' needs the ULID of the object it numbers within")
		}
		parts = append(parts, subScope+":"+objectId)
	}
	if len(parts) == 0 {
		return def.GetConfigurationScope(), nil
	}
	return strings.Join(parts, "/"), nil
}

func invalidScopeObject(message string) *ft.ClientErrors {
	cErrs := singleError(newSeqValidation("scope_object_id", ErrKeySeqInvalidInput, message))
	return &cErrs
}

// looksLikeUlid accepts what a ulid field accepts (go.bryk.io ulid.Parse), so an id the rest of the
// system stores, including seeded ones outside the strict Crockford alphabet, can scope a counter.
func looksLikeUlid(value string) bool {
	if len(value) != model.MODEL_RULE_ULID_LENGTH {
		return false
	}
	_, err := ulid.Parse(value)
	return err == nil
}

// FormatNumber renders the issued value (BR §5.1): prefix + number + suffix, where the number is the
// zero-padded decimal or, for FF1, the pseudorandom code. A consumer-formatted definition gets the
// number part only.
func FormatNumber(
	def *models.SequenceDefinition, counter *itSeq.IncrementedCounter, vars seqalgo.TemplateVars,
) (string, *ft.ClientErrors, error) {
	core, cErrs, err := formatCore(def, counter)
	if err != nil || cErrs != nil || def.GetFormatOwner() == models.SeqFormatOwnerConsumer {
		return core, cErrs, err
	}
	location, err := time.LoadLocation(def.GetTimezone())
	if err != nil {
		return "", nil, errors.Wrap(err, "FormatNumber.LoadLocation")
	}
	vars.Location = location
	prefix, prefixErr := seqalgo.RenderTemplate(def.GetPrefixTemplate(), vars)
	suffix, suffixErr := seqalgo.RenderTemplate(def.GetSuffixTemplate(), vars)
	vErrs := ft.NewClientErrors()
	appendTemplateError(models.SeqDefFieldPrefixTemplate, prefixErr, vErrs)
	appendTemplateError(models.SeqDefFieldSuffixTemplate, suffixErr, vErrs)
	formatted := prefix + core + suffix
	if vErrs.Count() == 0 && len(formatted) > seqMaxFormattedLength {
		vErrs.Append(*newSeqValidation(models.SeqDefFieldSuffixTemplate, ErrKeySeqInvalidTemplate,
			"the code exceeds {{max}} characters", map[string]any{"max": seqMaxFormattedLength}))
	}
	if vErrs.Count() > 0 {
		return "", vErrs, nil
	}
	return formatted, nil, nil
}

func formatCore(
	def *models.SequenceDefinition, counter *itSeq.IncrementedCounter,
) (string, *ft.ClientErrors, error) {
	if def.GetAlgorithm() != models.SeqAlgorithmFf1Fpe {
		return fmt.Sprintf("%0*d", int(def.GetPadding()), counter.LastNumber), nil, nil
	}
	tweak := models.JsonMapString(counter.AlgorithmParams, models.SeqAlgorithmParamTweak)
	code, err := seqalgo.Pseudorandom(counter.LastNumber, def.GetInitialValue(),
		def.GetAlgorithmParam(models.SeqAlgorithmParamKey), tweak)
	if errors.Is(err, seqalgo.ErrCapacityExceeded) {
		cErrs := singleError(newSeqViolation("definition_id", ErrKeySeqCapacityExceeded,
			"the pseudorandom sequence has used every code up to its maximum length"))
		return "", &cErrs, nil
	}
	return code, nil, errors.Wrap(err, "FormatNumber.Pseudorandom")
}

func newResultData(
	def *models.SequenceDefinition, req nextValueRequest, done itSeq.CompleteIssuanceParam,
	scopeKey string, periodKey string,
) itSeq.NextValueResultData {
	return itSeq.NextValueResultData{
		IssuanceId: done.Id, DefinitionId: def.GetId(), CounterId: done.CounterId,
		ScopeKey: scopeKey, PeriodKey: periodKey,
		AllocatedValue: done.AllocatedValue, FormattedValue: done.FormattedValue,
		EffectiveAt:   model.WrapModelDateTime(done.EffectiveAt).String(),
		IssuedAt:      model.WrapModelDateTime(done.IssuedAt).String(),
		ExpiresAt:     model.WrapModelDateTime(done.ExpiresAt).String(),
		IdempotencyId: req.idempotencyId, OperationKey: req.operationKey,
	}
}

// replayedData answers from the stored row, never re-rendering, so a replay keeps its period.
func replayedData(
	issuance dmodel.DynamicFields, counter dmodel.DynamicFields, req nextValueRequest,
) itSeq.NextValueResultData {
	allocated, _ := models.FieldInt64(issuance, models.SeqIssFieldAllocatedValue)
	return itSeq.NextValueResultData{
		IssuanceId:     model.Id(models.FieldString(issuance, models.SeqIssFieldId)),
		DefinitionId:   req.cmd.DefinitionId,
		CounterId:      model.Id(models.FieldString(issuance, models.SeqIssFieldCounterId)),
		ScopeKey:       models.FieldString(counter, models.SeqCtrFieldScopeKey),
		PeriodKey:      models.FieldString(counter, models.SeqCtrFieldPeriodKey),
		AllocatedValue: allocated,
		FormattedValue: models.FieldString(issuance, models.SeqIssFieldFormattedValue),
		EffectiveAt:    dateTimeString(issuance, models.SeqIssFieldEffectiveAt),
		IssuedAt:       dateTimeString(issuance, models.SeqIssFieldIssuedAt),
		ExpiresAt:      dateTimeString(issuance, models.SeqIssFieldExpiresAt),
		IdempotencyId:  req.idempotencyId,
		OperationKey:   req.operationKey,
		Replayed:       true,
	}
}

func dateTimeString(fields dmodel.DynamicFields, key string) string {
	value := fields.GetModelDateTime(key)
	if value == nil {
		return ""
	}
	return value.String()
}

func optionalId(value string) *model.Id {
	if value == "" {
		return nil
	}
	id := model.Id(value)
	return &id
}

// actingTenantId reads the tenant from the request's domain constraints, where the multi-tenant
// layer puts it; empty in the single-tenant binary.
func actingTenantId(ctx corectx.Context) string {
	constraints := ctx.GetDomainConstraints()
	if constraints == nil || constraints[seqTenantKeyField] == nil {
		return ""
	}
	return fmt.Sprint(constraints[seqTenantKeyField])
}

func actingPrincipalId(ctx corectx.Context) string {
	perms := ctx.GetPermissions()
	if perms.UserId != "" {
		return string(perms.UserId)
	}
	return string(perms.Principal.Id)
}
