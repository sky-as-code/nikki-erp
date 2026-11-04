package repository

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

func TestBuildIncrementCounterSql_TenantAndOverflowGuard(t *testing.T) {
	tenant := tenantScope{column: "tenant_id", value: "T1"}
	query, args, err := buildIncrementCounterSql("essential_sequence_counters", tenant, itSeq.IncrementCounterParam{
		NewId: "C1", DefinitionId: "D1", ScopeKey: "tenant:T1/org:O1", PeriodKey: "NONE",
		InitialValue: 1, Increment: 1, AlgorithmParams: dmodel.DynamicFields{"tweak": "ab"},
	}, "U1")
	assert.NoError(t, err)
	assert.Contains(t, query, "ON CONFLICT (tenant_id, definition_id, scope_key, period_key)")
	assert.Contains(t, query, "WHERE essential_sequence_counters.last_number <= ")
	assert.Contains(t, query, "RETURNING id, last_number")
	assert.Equal(t, strings.Count(query, "$"), len(args)+1, "the increment placeholder is used twice")
	assert.Contains(t, args, "T1")
	assert.Contains(t, args, `{"tweak":"ab"}`)
}

func TestBuildIncrementCounterSql_NoTenantKey(t *testing.T) {
	query, args, err := buildIncrementCounterSql("essential_sequence_counters", tenantScope{},
		itSeq.IncrementCounterParam{NewId: "C1", DefinitionId: "D1", ScopeKey: "org:O1", PeriodKey: "NONE",
			InitialValue: 1, Increment: 1}, nil)
	assert.NoError(t, err)
	assert.Contains(t, query, "ON CONFLICT (definition_id, scope_key, period_key)")
	assert.NotContains(t, query, "tenant_id")
	assert.Nil(t, args[7], "increment counters carry no algorithm_params")
}

func TestBuildReserveIssuanceSql(t *testing.T) {
	query, args := buildReserveIssuanceSql("essential_sequence_issuances", tenantScope{column: "tenant_id", value: "T1"},
		itSeq.ReserveIssuanceParam{NewId: "I1", DefinitionId: "D1", IdempotencyId: "u", OperationKey: "default"}, nil)
	assert.Contains(t, query, "ON CONFLICT (tenant_id, definition_id, idempotency_id, operation_key) DO NOTHING")
	assert.Contains(t, query, "RETURNING id")
	assert.Equal(t, strings.Count(query, "$"), len(args))
}
