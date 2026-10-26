package repository

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
)

func TestBuildReserveSql_ChecksTheCapInTheSameStatement(t *testing.T) {
	capBytes := int64(1000)
	owner := models.ConsumptionOwner{Type: "tenant", Id: "T1"}
	query, args := buildReserveSql("itstorage_consumptions", owner, 42, &capBytes, "tenant_id", "T1")

	assert.Contains(t, query, "reserved_bytes = reserved_bytes + $")
	assert.Contains(t, query, "used_bytes + reserved_bytes + $")
	assert.Contains(t, query, "owner_type = $")
	assert.Contains(t, query, "tenant_id = $")
	assert.Contains(t, args, int64(42))
	assert.Contains(t, args, int64(1000))
}

func TestBuildReserveSql_WithoutCap(t *testing.T) {
	owner := models.ConsumptionOwner{Type: "user", Id: "U1"}
	query, _ := buildReserveSql("itstorage_consumptions", owner, 42, nil, "", "")

	assert.NotContains(t, query, "<=")
	assert.NotContains(t, query, "tenant_id")
}

func TestBuildEnsureSql_DoesNothingOnConflict(t *testing.T) {
	owners := []models.ConsumptionOwner{
		{Type: "tenant", Id: "T1", InitialUsedBytes: 500},
		{Type: "user", Id: "U1"},
	}
	query, args, err := buildEnsureSql("itstorage_consumptions", owners, "tenant_id", "T1")

	assert.NoError(t, err)
	assert.True(t, strings.HasSuffix(query, "ON CONFLICT DO NOTHING"), query)
	assert.Contains(t, query, "tenant_id")
	assert.Len(t, args, 16)
	assert.Equal(t, int64(500), args[3])
}

func TestBuildSetStatusSql_ToInUseSkipsReleasedFiles(t *testing.T) {
	query, args := buildSetStatusSql("itstorage_files", []model.Id{"F1", "F2"}, "pending", "in_use", "", "")

	assert.Contains(t, query, "in_use_at IS NULL")
	assert.Contains(t, query, "RETURNING id, object_key")
	assert.Contains(t, args, "pending")
	assert.Contains(t, args, "in_use")
}

func TestBuildSetStatusSql_ToPendingKeepsInUseAt(t *testing.T) {
	query, _ := buildSetStatusSql("itstorage_files", []model.Id{"F1"}, "in_use", "pending", "tenant_id", "T1")

	assert.NotContains(t, query, "in_use_at")
	assert.Contains(t, query, "tenant_id = $")
}
