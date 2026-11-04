package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SequenceDefinitionSchemaName = "essential_sequence_definition"

	SeqDefFieldId                    = basemodel.FieldId
	SeqDefFieldEtag                  = basemodel.FieldEtag
	SeqDefFieldOrgId                 = basemodel.FieldOrgId
	SeqDefFieldIsArchived            = basemodel.FieldIsArchived
	SeqDefFieldConfigurationScope    = "configuration_scope"
	SeqDefFieldOwnerModule           = "owner_module"
	SeqDefFieldCode                  = "code"
	SeqDefFieldName                  = "name"
	SeqDefFieldDescription           = "description"
	SeqDefFieldSubScope              = "sub_scope"
	SeqDefFieldScopeType             = "scope_type"
	SeqDefFieldFormatOwner           = "format_owner"
	SeqDefFieldAlgorithm             = "algorithm"
	SeqDefFieldAlgorithmParams       = "algorithm_params"
	SeqDefFieldPrefixTemplate        = "prefix_template"
	SeqDefFieldSuffixTemplate        = "suffix_template"
	SeqDefFieldPadding               = "padding"
	SeqDefFieldIncrement             = "increment"
	SeqDefFieldInitialValue          = "initial_value"
	SeqDefFieldResetPeriod           = "reset_period"
	SeqDefFieldTimezone              = "timezone"
	SeqDefFieldIssuanceRetentionDays = "issuance_retention_days"
	SeqDefFieldFirstIssuedAt         = "first_issued_at"
)

// Stored enum values. The API spells FF1_FPE as "pseudorandom" (see the application service).
const (
	SeqConfigScopeTenant = "tenant"
	SeqConfigScopeOrg    = "org"

	SeqFormatOwnerSequence = "sequence"
	SeqFormatOwnerConsumer = "consumer"

	SeqAlgorithmIncrement = "increment"
	SeqAlgorithmFf1Fpe    = "FF1_FPE"

	SeqResetNever   = "never"
	SeqResetYearly  = "yearly"
	SeqResetMonthly = "monthly"
	SeqResetDaily   = "daily"

	// SeqAlgorithmParamKey is the FF1 key (hex of 32 random bytes) in a definition's algorithm_params.
	SeqAlgorithmParamKey = "key"
	// SeqAlgorithmParamTweak is the FF1 tweak (hex of 16 random bytes) in a counter's algorithm_params.
	SeqAlgorithmParamTweak = "tweak"

	SeqDefaultPadding       = int32(5)
	SeqDefaultRetentionDays = int32(90)
	SeqDefaultTimezone      = "UTC"
	SeqDefaultOperationKey  = "default"
)

//go:embed sequence_definition.json
var sequenceDefinitionSchemaJson string

func SequenceDefinitionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(sequenceDefinitionSchemaJson)
}

// SequenceDefinition reads a definition row. Its getters tolerate the shapes a value takes before
// schema validation (a JSON number arrives as float64), so a rule can read raw input too.
type SequenceDefinition struct {
	fields dmodel.DynamicFields
}

func NewSequenceDefinitionFrom(src dmodel.DynamicFields) *SequenceDefinition {
	if src == nil {
		src = dmodel.DynamicFields{}
	}
	return &SequenceDefinition{fields: src}
}

func (this SequenceDefinition) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this SequenceDefinition) GetId() model.Id {
	return model.Id(FieldString(this.fields, SeqDefFieldId))
}

func (this SequenceDefinition) GetOrgId() string {
	return FieldString(this.fields, SeqDefFieldOrgId)
}

func (this SequenceDefinition) GetConfigurationScope() string {
	return FieldString(this.fields, SeqDefFieldConfigurationScope)
}

func (this SequenceDefinition) GetOwnerModule() string {
	return FieldString(this.fields, SeqDefFieldOwnerModule)
}

func (this SequenceDefinition) GetCode() string {
	return FieldString(this.fields, SeqDefFieldCode)
}

func (this SequenceDefinition) GetSubScope() string {
	return FieldString(this.fields, SeqDefFieldSubScope)
}

func (this SequenceDefinition) GetScopeType() string {
	return FieldString(this.fields, SeqDefFieldScopeType)
}

func (this SequenceDefinition) GetFormatOwner() string {
	return FieldStringOr(this.fields, SeqDefFieldFormatOwner, SeqFormatOwnerSequence)
}

func (this SequenceDefinition) GetAlgorithm() string {
	return FieldStringOr(this.fields, SeqDefFieldAlgorithm, SeqAlgorithmIncrement)
}

func (this SequenceDefinition) GetPrefixTemplate() string {
	return FieldString(this.fields, SeqDefFieldPrefixTemplate)
}

func (this SequenceDefinition) GetSuffixTemplate() string {
	return FieldString(this.fields, SeqDefFieldSuffixTemplate)
}

func (this SequenceDefinition) GetPadding() int64 {
	value, _ := FieldInt64(this.fields, SeqDefFieldPadding)
	return value
}

func (this SequenceDefinition) GetIncrement() int64 {
	value, _ := FieldInt64(this.fields, SeqDefFieldIncrement)
	return value
}

func (this SequenceDefinition) GetInitialValue() int64 {
	value, _ := FieldInt64(this.fields, SeqDefFieldInitialValue)
	return value
}

func (this SequenceDefinition) GetResetPeriod() string {
	return FieldStringOr(this.fields, SeqDefFieldResetPeriod, SeqResetNever)
}

func (this SequenceDefinition) GetTimezone() string {
	return FieldStringOr(this.fields, SeqDefFieldTimezone, SeqDefaultTimezone)
}

func (this SequenceDefinition) GetRetentionDays() int64 {
	value, found := FieldInt64(this.fields, SeqDefFieldIssuanceRetentionDays)
	if !found {
		return int64(SeqDefaultRetentionDays)
	}
	return value
}

func (this SequenceDefinition) IsArchived() bool {
	archived, _ := this.fields[SeqDefFieldIsArchived].(bool)
	return archived
}

func (this SequenceDefinition) IsFirstIssued() bool {
	return this.fields[SeqDefFieldFirstIssuedAt] != nil
}

// GetAlgorithmParam reads one key of algorithm_params, empty when absent.
func (this SequenceDefinition) GetAlgorithmParam(key string) string {
	return JsonMapString(this.fields[SeqDefFieldAlgorithmParams], key)
}
