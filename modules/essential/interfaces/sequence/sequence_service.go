package sequence

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// API spellings of the algorithm. The database stores AlgorithmPseudorandom as FF1_FPE.
const (
	AlgorithmIncrement    = "increment"
	AlgorithmPseudorandom = "pseudorandom"
)

// SequenceService is the port feature modules consume through their own interfaces/external. It
// asserts no permission: a module reaches it only from code that already authorized its caller.
// NextValue joins the transaction ctx carries, so a caller's rollback releases the number; a caller
// that sees client errors must roll back, since the reservation is already written.
type SequenceService interface {
	UpsertDefinition(ctx corectx.Context, cmd UpsertDefinitionCommand) (*UpsertDefinitionResult, error)
	NextValue(ctx corectx.Context, cmd NextValueCommand) (*NextValueResult, error)
}

// SequenceDomainService carries the same two operations below the application layer. Its commands use
// the stored vocabulary (algorithm FF1_FPE) and a resolved idempotency id.
type SequenceDomainService interface {
	UpsertDefinition(ctx corectx.Context, cmd UpsertDefinitionCommand) (*UpsertDefinitionResult, error)
	NextValue(ctx corectx.Context, cmd NextValueCommand) (*NextValueResult, error)
}

// UpsertDefinitionCommand registers a definition, or returns the existing one with the same
// (configuration_scope, org, code) unchanged. OrgId is nil for configuration_scope "tenant".
// Names are candidate translations; the definition stores the one in the org's system_language
// (DefaultLanguageCode for a tenant definition).
type UpsertDefinitionCommand struct {
	ConfigurationScope    string         `json:"configuration_scope"`
	OrgId                 *model.Id      `json:"org_id"`
	OwnerModule           string         `json:"owner_module"`
	Code                  string         `json:"code"`
	Names                 model.LangJson `json:"names"`
	Description           *string        `json:"description"`
	SubScope              *string        `json:"sub_scope"`
	FormatOwner           *string        `json:"format_owner"`
	Algorithm             *string        `json:"algorithm"`
	PrefixTemplate        *string        `json:"prefix_template"`
	SuffixTemplate        *string        `json:"suffix_template"`
	Padding               *int32         `json:"padding"`
	Increment             *int64         `json:"increment"`
	InitialValue          *int64         `json:"initial_value"`
	ResetPeriod           *string        `json:"reset_period"`
	Timezone              *string        `json:"timezone"`
	IssuanceRetentionDays *int32         `json:"issuance_retention_days"`
}

type UpsertDefinitionResultData struct {
	DefinitionId model.Id `json:"definition_id"`
	Created      bool     `json:"created"`
}

type UpsertDefinitionResult = dyn.OpResult[UpsertDefinitionResultData]

// NextValueCommand asks for one number. IdempotencyId falls back to the one ctx carries, then to a
// fresh one; OperationKey defaults to "default". EffectiveAt (RFC3339) defaults to the acceptance time.
type NextValueCommand struct {
	DefinitionId  model.Id  `json:"definition_id"`
	OrgId         *model.Id `json:"org_id"`
	CallerModule  string    `json:"caller_module"`
	ScopeObjectId *model.Id `json:"scope_object_id"`
	IdempotencyId *string   `json:"idempotency_id"`
	OperationKey  *string   `json:"operation_key"`
	EffectiveAt   *string   `json:"effective_at"`
}

type NextValueResultData struct {
	IssuanceId     model.Id `json:"issuance_id"`
	DefinitionId   model.Id `json:"definition_id"`
	CounterId      model.Id `json:"counter_id"`
	ScopeKey       string   `json:"scope_key"`
	PeriodKey      string   `json:"period_key"`
	AllocatedValue int64    `json:"allocated_value,string"`
	FormattedValue string   `json:"formatted_value"`
	EffectiveAt    string   `json:"effective_at"`
	IssuedAt       string   `json:"issued_at"`
	ExpiresAt      string   `json:"expires_at"`
	IdempotencyId  string   `json:"idempotency_id"`
	OperationKey   string   `json:"operation_key"`
	Replayed       bool     `json:"replayed"`
}

type NextValueResult = dyn.OpResult[NextValueResultData]
