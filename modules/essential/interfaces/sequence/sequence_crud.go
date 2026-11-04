package sequence

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// LockMode is the row lock a definition read takes. Share lets issuances run side by side; update
// serializes an admin change or archive against them.
type LockMode string

const (
	LockModeShare  = LockMode("SHARE")
	LockModeUpdate = LockMode("UPDATE")
)

type SequenceDefinitionRepository interface {
	dynamicresource.CrudRepository

	// LockDefinition takes the row lock and reports whether the row exists in the caller's tenant.
	LockDefinition(ctx corectx.Context, id model.Id, mode LockMode) (bool, error)
	// StampFirstIssued sets first_issued_at once; it reports whether this call set it. The UPDATE is
	// also the row's exclusive lock for the rest of the transaction.
	StampFirstIssued(ctx corectx.Context, id model.Id, at time.Time) (bool, error)
	// LockRegistration serializes UpsertDefinition calls on one registration key until the
	// transaction ends (a transaction-scoped advisory lock).
	LockRegistration(ctx corectx.Context, key string) error
	// FindIdByCode finds the definition registered under code for the org, or for the tenant when
	// orgId is nil. Empty when there is none.
	FindIdByCode(ctx corectx.Context, code string, orgId *model.Id) (model.Id, error)
}

// IncrementCounterParam is one atomic find-or-create-and-add on a counter.
type IncrementCounterParam struct {
	NewId           model.Id
	DefinitionId    model.Id
	OrgId           *model.Id
	ScopeObjectId   *model.Id
	ScopeKey        string
	PeriodKey       string
	InitialValue    int64
	Increment       int64
	AlgorithmParams dmodel.DynamicFields
}

// IncrementedCounter is the counter row after the increment.
type IncrementedCounter struct {
	Id              model.Id
	LastNumber      int64
	AlgorithmParams []byte
}

type SequenceCounterRepository interface {
	dynamicresource.CrudRepository

	// IncrementCounter returns nil when the increment would overflow int64 (SEQUENCE_EXHAUSTED).
	IncrementCounter(ctx corectx.Context, param IncrementCounterParam) (*IncrementedCounter, error)
}

// ReserveIssuanceParam is the request row written before any counter is touched.
type ReserveIssuanceParam struct {
	NewId              model.Id
	DefinitionId       model.Id
	OrgId              *model.Id
	IdempotencyId      string
	OperationKey       string
	RequestFingerprint string
	CallerModule       string
	CallerPrincipalId  *model.Id
	AcceptedAt         time.Time
}

// CompleteIssuanceParam fills in the result columns of a reserved issuance.
type CompleteIssuanceParam struct {
	Id             model.Id
	CounterId      model.Id
	EffectiveAt    time.Time
	AllocatedValue int64
	FormattedValue string
	IssuedAt       time.Time
	ExpiresAt      time.Time
}

type SequenceIssuanceRepository interface {
	dynamicresource.CrudRepository

	// ReserveIssuance reports false when the request key already exists; Postgres has then waited
	// for the transaction that holds it.
	ReserveIssuance(ctx corectx.Context, param ReserveIssuanceParam) (bool, error)
	CompleteIssuance(ctx corectx.Context, param CompleteIssuanceParam) error
}

type SequenceDefinitionDomainService interface {
	dynamicresource.CrudDomainService
}

type SequenceCounterDomainService interface {
	dynamicresource.CrudDomainService
}

type SequenceIssuanceDomainService interface {
	dynamicresource.CrudDomainService
}

type SequenceDefinitionApplicationService interface {
	dynamicresource.CrudApplicationService

	// Preview renders a sample number for an unsaved configuration. It writes nothing.
	Preview(ctx corectx.Context, cmd PreviewCommand) (*PreviewResult, error)
}

type SequenceCounterApplicationService interface {
	dynamicresource.CrudApplicationService
}

type SequenceIssuanceApplicationService interface {
	dynamicresource.CrudApplicationService
}

// PreviewCommand is the configuration to render plus optional sample number and time; org_id
// names the org whose definitions the caller may read.
type PreviewCommand = dmodel.DynamicFields

type PreviewResultData struct {
	FormattedValue string `json:"formatted_value"`
	PeriodKey      string `json:"period_key"`
	SampleNumber   int64  `json:"sample_number,string"`
	SampleAt       string `json:"sample_at"`
	// IsIllustration is always true: a preview reserves nothing.
	IsIllustration bool `json:"is_illustration"`
}

type PreviewResult = dyn.OpResult[PreviewResultData]

type (
	CreateSequenceDefinitionCommand = dynamicresource.CreateCommand
	UpdateSequenceDefinitionCommand = dynamicresource.UpdateCommand
	SetSequenceArchivedCommand      = dynamicresource.SetArchivedCommand
	SearchSequenceQuery             = dynamicresource.SearchQuery
	GetSequenceByIdQuery            = dynamicresource.GetByIdQuery
)
