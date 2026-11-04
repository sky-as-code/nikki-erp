package app

import (
	"encoding/json"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/services"
	seqalgo "github.com/sky-as-code/nikki-erp/modules/essential/domain/services/sequence"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

const (
	previewFieldSampleNumber = "sample_number"
	previewFieldSampleAt     = "sample_at"
)

// Preview renders the code a configuration would produce for a sample number and time (BR §4.2). It
// needs read permission, touches no counter or issuance and reserves nothing. A pseudorandom preview
// uses a throwaway key, so it shows the shape of the code, never a code the definition will issue.
func (this *SequenceDefinitionApplicationServiceImpl) Preview(
	ctx corectx.Context, cmd itSeq.PreviewCommand,
) (*itSeq.PreviewResult, error) {
	_, cErrs := dynamicresource.AssertActionPermission(ctx, this, cmd,
		dynamicresource.ActionPermission{ActionCode: dynamicresource.PermissionRead})
	if cErrs != nil {
		return &itSeq.PreviewResult{ClientErrors: *cErrs}, nil
	}
	def, sampleAt, cErrs := previewDefinition(cmd)
	if cErrs != nil {
		return &itSeq.PreviewResult{ClientErrors: *cErrs}, nil
	}
	counter, err := previewCounter(def, cmd)
	if err != nil {
		return nil, err
	}
	period, err := seqalgo.PeriodOf(sampleAt, def.GetTimezone(), def.GetResetPeriod())
	if err != nil {
		return nil, err
	}
	formatted, cErrs, err := services.FormatNumber(def, counter, seqalgo.TemplateVars{
		EffectiveAt: sampleAt, AcceptedAt: sampleAt, RangeStart: period.Start, PeriodKey: period.Key,
	})
	if err != nil || cErrs != nil {
		return previewFailure(cErrs, err)
	}
	return &itSeq.PreviewResult{HasData: true, Data: itSeq.PreviewResultData{
		FormattedValue: formatted, PeriodKey: period.Key, SampleNumber: counter.LastNumber,
		SampleAt: model.WrapModelDateTime(sampleAt).String(), IsIllustration: true,
	}}, nil
}

// previewDefinition fills the defaults a create would and validates the result like one.
func previewDefinition(cmd dmodel.DynamicFields) (*models.SequenceDefinition, time.Time, *ft.ClientErrors) {
	fields := dmodel.DynamicFields{
		models.SeqDefFieldFormatOwner: models.SeqFormatOwnerSequence,
		models.SeqDefFieldAlgorithm:   models.SeqAlgorithmIncrement,
		models.SeqDefFieldPadding:     int64(models.SeqDefaultPadding),
		models.SeqDefFieldIncrement:   int64(1), models.SeqDefFieldInitialValue: int64(1),
		models.SeqDefFieldResetPeriod: models.SeqResetNever, models.SeqDefFieldTimezone: models.SeqDefaultTimezone,
	}
	for _, key := range []string{
		models.SeqDefFieldFormatOwner, models.SeqDefFieldAlgorithm, models.SeqDefFieldPrefixTemplate,
		models.SeqDefFieldSuffixTemplate, models.SeqDefFieldPadding, models.SeqDefFieldIncrement,
		models.SeqDefFieldInitialValue, models.SeqDefFieldResetPeriod, models.SeqDefFieldTimezone,
	} {
		if value, ok := cmd[key]; ok && value != nil && value != "" {
			fields[key] = value
		}
	}
	mapAlgorithmInput(fields)
	if models.FieldString(fields, models.SeqDefFieldAlgorithm) == models.SeqAlgorithmFf1Fpe {
		fields[models.SeqDefFieldPadding] = int64(0)
	}
	vErrs := ft.NewClientErrors()
	def := models.NewSequenceDefinitionFrom(fields)
	services.ValidateDefinitionConfig(def, vErrs)
	sampleAt := previewSampleAt(cmd, vErrs)
	if vErrs.Count() > 0 {
		return nil, sampleAt, vErrs
	}
	return def, sampleAt, nil
}

func previewSampleAt(cmd dmodel.DynamicFields, vErrs *ft.ClientErrors) time.Time {
	raw := models.FieldString(cmd, previewFieldSampleAt)
	if raw == "" {
		return time.Now().UTC()
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		vErrs.Append(*ft.NewValidationError(previewFieldSampleAt, services.ErrKeySeqInvalidInput,
			"sample_at must be an RFC 3339 timestamp with a time zone"))
	}
	return parsed.UTC()
}

// previewCounter is a counter that never existed: the sample number, and for FF1 a throwaway key and tweak.
func previewCounter(def *models.SequenceDefinition, cmd dmodel.DynamicFields) (*itSeq.IncrementedCounter, error) {
	sample, ok := models.FieldInt64(cmd, previewFieldSampleNumber)
	if !ok || sample < def.GetInitialValue() {
		sample = def.GetInitialValue()
	}
	counter := &itSeq.IncrementedCounter{LastNumber: sample}
	if def.GetAlgorithm() != models.SeqAlgorithmFf1Fpe {
		return counter, nil
	}
	key, err := seqalgo.NewKeyHex()
	if err != nil {
		return nil, err
	}
	tweak, err := seqalgo.NewTweakHex()
	if err != nil {
		return nil, err
	}
	def.GetFieldData()[models.SeqDefFieldAlgorithmParams] = map[string]any{models.SeqAlgorithmParamKey: key}
	counter.AlgorithmParams, err = json.Marshal(map[string]string{models.SeqAlgorithmParamTweak: tweak})
	return counter, err
}

func previewFailure(cErrs *ft.ClientErrors, err error) (*itSeq.PreviewResult, error) {
	if err != nil {
		return nil, err
	}
	return &itSeq.PreviewResult{ClientErrors: *cErrs}, nil
}
