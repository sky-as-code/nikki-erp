package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildAddSizeSql_WalksTheParentChain(t *testing.T) {
	query, args := buildAddSizeSql("itstorage_spaces", "S1", 42, "tenant_id", "T1")

	assert.Equal(t, "WITH RECURSIVE chain(id, parent_id) AS ("+
		"SELECT id, parent_id FROM itstorage_spaces WHERE id = $1 AND tenant_id = $2"+
		" UNION ALL SELECT s.id, s.parent_id FROM itstorage_spaces s JOIN chain ON s.id = chain.parent_id"+
		") UPDATE itstorage_spaces SET size_bytes = GREATEST(size_bytes + $3, 0)"+
		" WHERE id IN (SELECT id FROM chain) AND tenant_id = $4", query)
	assert.Equal(t, []any{"S1", "T1", int64(42), "T1"}, args)
}

func TestBuildAddSizeSql_WithoutTenant(t *testing.T) {
	query, args := buildAddSizeSql("itstorage_spaces", "S1", -7, "", "")

	assert.NotContains(t, query, "tenant_id")
	assert.Equal(t, []any{"S1", int64(-7)}, args)
}

func TestBuildRewritePathsSql_UsesALiteralPrefix(t *testing.T) {
	query, args := buildRewritePathsSql("itstorage_spaces", "root/old_name", "root/new", "tenant_id", "T1")

	assert.Equal(t, "UPDATE itstorage_spaces SET path = $1 || substr(path, $2)"+
		" WHERE substr(path, 1, $3) = $4 AND tenant_id = $5", query)
	assert.Equal(t, []any{"root/new", 14, 14, "root/old_name/", "T1"}, args)
}

func TestBuildListSubtreeSql_DeepestFirst(t *testing.T) {
	query, args := buildListSubtreeSql("itstorage_spaces", "root/org", "", "")

	assert.Equal(t, "SELECT id FROM itstorage_spaces WHERE (path = $1 OR substr(path, 1, $2) = $3)"+
		" ORDER BY length(path) DESC", query)
	assert.Equal(t, []any{"root/org", 9, "root/org/"}, args)
}
