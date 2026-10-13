package advanced

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// `*` projects the root's own columns as one wildcard, however many there are, and leaves every
// other entry (computed fields, edge paths) to render beside it as usual.

func TestWildcard_AloneIsPlainStar(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("*")})

	assert.Equal(t, `SELECT * FROM "adv_templates"`, sql)
}

func TestWildcard_WithSqlComputedFieldAliasesTheRoot(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("*", "on_hand_quantity")})

	assert.True(t, strings.HasPrefix(sql, `SELECT t0.*, c0.v AS "on_hand_quantity" FROM "adv_templates" AS t0`), sql)
	assert.Contains(t, sql, "LEFT JOIN LATERAL")
}

func TestWildcard_DropsRootColumnsItAlreadyCovers(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaTemplate, nil, orm.SqlSelectGraphOpts{Columns: cols("*", "id", "code")})

	assert.Equal(t, `SELECT * FROM "adv_templates"`, sql)
}

func TestWildcard_WithToOneEdgePath(t *testing.T) {
	fx := newFixture(t)
	sql := advSelect(t, fx, schemaVariant, nil, orm.SqlSelectGraphOpts{Columns: cols("*", "template.code")})

	assert.True(t, strings.HasPrefix(sql, `SELECT t0.*, t1."id" AS "template.id", t1."code" AS "template.code" FROM`), sql)
	assert.NotContains(t, sql, `t0."id"`)
}

func TestWildcard_SkipsRootOrderRefsUnderDistinct(t *testing.T) {
	fx := newFixture(t)
	graph := dmodel.NewSearchGraph().OrderBy("code", dmodel.Asc)
	sql := advSelect(t, fx, schemaTemplate, graph, orm.SqlSelectGraphOpts{Columns: cols("*", "on_hand_quantity")})

	assert.True(t, strings.HasPrefix(sql, `SELECT t0.*, c0.v AS "on_hand_quantity" FROM`), sql)
	assert.Contains(t, sql, `ORDER BY t0."code" ASC`)
}
