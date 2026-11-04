package services

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	seqalgo "github.com/sky-as-code/nikki-erp/modules/essential/domain/services/sequence"
)

// ValidateDefinitionConfig checks what the schema cannot: padding per algorithm, the time zone, the
// templates (syntax, digit boundary SEQ-UNIQ-02, period marker SEQ-TIME-05) and the worst-case length.
func ValidateDefinitionConfig(def *models.SequenceDefinition, vErrs *ft.ClientErrors) {
	assertPadding(def, vErrs)
	if _, err := time.LoadLocation(def.GetTimezone()); err != nil {
		vErrs.Append(*newSeqValidation(models.SeqDefFieldTimezone, ErrKeySeqInvalidTimezone,
			"'{{timezone}}' is not an IANA time zone", map[string]any{"timezone": def.GetTimezone()}))
	}
	assertTemplates(def, vErrs)
}

func assertPadding(def *models.SequenceDefinition, vErrs *ft.ClientErrors) {
	padding := def.GetPadding()
	if def.GetAlgorithm() == models.SeqAlgorithmFf1Fpe {
		if padding != 0 {
			vErrs.Append(*newSeqValidation(models.SeqDefFieldPadding, ErrKeySeqPadding,
				"a pseudorandom definition has no padding; it must be 0"))
		}
		return
	}
	if padding < 1 || padding > seqMaxPadding {
		vErrs.Append(*newSeqValidation(models.SeqDefFieldPadding, ErrKeySeqPadding,
			"padding must be between {{min}} and {{max}}", map[string]any{"min": 1, "max": seqMaxPadding}))
	}
}

func assertTemplates(def *models.SequenceDefinition, vErrs *ft.ClientErrors) {
	prefix, suffix := def.GetPrefixTemplate(), def.GetSuffixTemplate()
	formatOwner, reset := def.GetFormatOwner(), def.GetResetPeriod()
	before := vErrs.Count()
	appendTemplateError(models.SeqDefFieldPrefixTemplate,
		seqalgo.ValidateTemplate(prefix, formatOwner, reset, true), vErrs)
	appendTemplateError(models.SeqDefFieldSuffixTemplate,
		seqalgo.ValidateTemplate(suffix, formatOwner, reset, false), vErrs)
	if vErrs.Count() > before {
		return
	}
	if formatOwner == models.SeqFormatOwnerSequence && reset != models.SeqResetNever {
		appendTemplateError(models.SeqDefFieldPrefixTemplate,
			seqalgo.ValidatePeriodMarker(prefix, suffix, reset), vErrs)
	}
	if len(prefix)+len(suffix)+seqMaxDigits > seqMaxFormattedLength {
		vErrs.Append(*newSeqValidation(models.SeqDefFieldSuffixTemplate, ErrKeySeqInvalidTemplate,
			"prefix, number and suffix may exceed {{max}} characters", map[string]any{"max": seqMaxFormattedLength}))
	}
}

// appendTemplateError turns an algorithm-package template error into the field's client error.
func appendTemplateError(field string, err error, vErrs *ft.ClientErrors) {
	if err == nil {
		return
	}
	vErrs.Append(*newSeqValidation(field, ErrKeySeqInvalidTemplate, "{{reason}}",
		map[string]any{"reason": err.Error()}))
}
