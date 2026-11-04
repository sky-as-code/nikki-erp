package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SequenceIssuanceSchemaName = "essential_sequence_issuance"

	SeqIssFieldId                 = basemodel.FieldId
	SeqIssFieldOrgId              = basemodel.FieldOrgId
	SeqIssFieldDefinitionId       = "definition_id"
	SeqIssFieldCounterId          = "counter_id"
	SeqIssFieldIdempotencyId      = "idempotency_id"
	SeqIssFieldOperationKey       = "operation_key"
	SeqIssFieldRequestFingerprint = "request_fingerprint"
	SeqIssFieldCallerModule       = "caller_module"
	SeqIssFieldCallerPrincipalId  = "caller_principal_id"
	SeqIssFieldAcceptedAt         = "accepted_at"
	SeqIssFieldEffectiveAt        = "effective_at"
	SeqIssFieldAllocatedValue     = "allocated_value"
	SeqIssFieldFormattedValue     = "formatted_value"
	SeqIssFieldIssuedAt           = "issued_at"
	SeqIssFieldExpiresAt          = "expires_at"
	SeqIssFieldCreatedAt          = basemodel.FieldCreatedAt
	SeqIssFieldCreatedBy          = "created_by"
)

//go:embed sequence_issuance.json
var sequenceIssuanceSchemaJson string

func SequenceIssuanceSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(sequenceIssuanceSchemaJson)
}
