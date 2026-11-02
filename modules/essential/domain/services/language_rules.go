package services

import (
	"encoding/json"
	"io"
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// The language rules follow Odoo's res.lang constraints (LANG-003), so a format a user copies from
// Odoo means the same thing here and the frontend can render every stored pattern.

// disallowedFormatDirectives are the strftime directives Odoo refuses: locale-dependent, non-portable
// or redundant. %y is deliberately allowed.
var disallowedFormatDirectives = []string{
	"%C", "%D", "%e", "%E", "%F", "%g", "%G", "%h", "%k", "%l", "%n",
	"%O", "%P", "%R", "%r", "%s", "%T", "%t", "%u", "%V", "%+",
}

var languageFormatFields = []string{
	models.LanguageFieldDateFormat, models.LanguageFieldTimeFormat, models.LanguageFieldShortTimeFormat,
}

const languageErrNamespace = "essential"

// languageView reads a field as it will be after the write: the input when it carries the field,
// the stored row otherwise. On create there is no stored row.
type languageView struct {
	input dmodel.DynamicFields
	found dmodel.DynamicFields
}

func newLanguageView(inputModel, foundModel *dynamicresource.DynamicEntity) languageView {
	view := languageView{input: dmodel.DynamicFields{}, found: dmodel.DynamicFields{}}
	if inputModel != nil {
		view.input = inputModel.GetFieldData()
	}
	if foundModel != nil {
		view.found = foundModel.GetFieldData()
	}
	return view
}

func (this languageView) sends(field string) bool {
	_, sent := this.input[field]
	return sent
}

func (this languageView) str(field string) string {
	if this.sends(field) {
		return derefStr(this.input.GetString(field))
	}
	return derefStr(this.found.GetString(field))
}

func derefStr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// assertLanguageFields checks the field-level rules on the fields this write sets, plus the
// separator pair as it will be stored.
func assertLanguageFields(view languageView, vErrs *ft.ClientErrors) {
	for _, field := range languageFormatFields {
		if view.sends(field) || len(view.found) == 0 {
			assertFormatDirectives(field, view.str(field), vErrs)
		}
	}
	if view.sends(models.LanguageFieldGrouping) && !isValidGrouping(view.str(models.LanguageFieldGrouping)) {
		vErrs.Append(*ft.NewValidationError(models.LanguageFieldGrouping,
			ft.ErrorKey("err_lang_grouping", languageErrNamespace),
			"grouping must be a JSON array of integers, such as [3,0]"))
	}
	decimal := view.str(models.LanguageFieldDecimalSeparator)
	thousands := view.str(models.LanguageFieldThousandsSeparator)
	if decimal != "" && decimal == thousands {
		field := models.LanguageFieldThousandsSeparator
		if view.sends(models.LanguageFieldDecimalSeparator) && !view.sends(field) {
			field = models.LanguageFieldDecimalSeparator
		}
		vErrs.Append(*ft.NewValidationError(field,
			ft.ErrorKey("err_lang_separators_equal", languageErrNamespace),
			"the decimal and thousands separators must differ"))
	}
}

func assertFormatDirectives(field string, format string, vErrs *ft.ClientErrors) {
	for _, directive := range disallowedFormatDirectives {
		if strings.Contains(format, directive) {
			vErrs.Append(*ft.NewValidationError(field,
				ft.ErrorKey("err_lang_format_directive", languageErrNamespace),
				"the format uses a directive that is not allowed",
				map[string]any{"directive": directive}))
			return
		}
	}
}

// isValidGrouping mirrors Odoo's _check_grouping: a JSON array whose every element is an integer.
// An empty array is valid (no grouping).
func isValidGrouping(grouping string) bool {
	trimmed := strings.TrimSpace(grouping)
	if !strings.HasPrefix(trimmed, "[") {
		return false
	}
	decoder := json.NewDecoder(strings.NewReader(trimmed))
	decoder.UseNumber()
	var items []any
	if err := decoder.Decode(&items); err != nil {
		return false
	}
	if _, err := decoder.Token(); err != io.EOF {
		return false
	}
	for _, item := range items {
		number, isNumber := item.(json.Number)
		if !isNumber {
			return false
		}
		if _, err := number.Int64(); err != nil {
			return false
		}
	}
	return true
}

// assertKeepsAnActiveLanguage refuses to deactivate or delete the organization's last active
// language: with none, nobody in it could read the interface (Odoo's _check_active).
func (this *LanguageDomainServiceImpl) assertKeepsAnActiveLanguage(
	ctx corectx.Context, orgId model.Id, selfId model.Id, vErrs *ft.ClientErrors,
) error {
	ids, err := this.activation.ActiveIds(ctx, orgId)
	if err != nil {
		return errors.Wrap(err, "assertKeepsAnActiveLanguage")
	}
	for _, id := range ids {
		if id != selfId {
			return nil
		}
	}
	vErrs.Append(*ft.NewBusinessViolation(models.LanguageFieldIsActive,
		ft.ErrorKey("err_lang_last_enabled", languageErrNamespace),
		"at least one language must stay active"))
	return nil
}

// assertNotTheDefaultLanguage refuses to disable or delete the language the acting organization
// reads by default: its users without a language of their own would lose the interface.
func (this *LanguageDomainServiceImpl) assertNotTheDefaultLanguage(
	ctx corectx.Context, view languageView, vErrs *ft.ClientErrors,
) error {
	defaultCode, err := this.defaultLocaleCode(ctx)
	if err != nil {
		return errors.Wrap(err, "assertNotTheDefaultLanguage")
	}
	if derefStr(view.found.GetString(models.LanguageFieldLocaleCode)) != defaultCode {
		return nil
	}
	vErrs.Append(*ft.NewBusinessViolation(models.LanguageFieldIsActive,
		ft.ErrorKey("err_lang_default_disabled", languageErrNamespace),
		"the default language cannot be disabled",
		map[string]any{"locale_code": defaultCode}))
	return nil
}
