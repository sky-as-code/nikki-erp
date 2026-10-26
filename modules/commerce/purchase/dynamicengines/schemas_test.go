package dynamicengines

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// TestMain registers the base mixins once for the whole package: ParseModelJson panics when a named
// mixin is missing, and CoreModule.RegisterModels normally does this at start-up.
func TestMain(m *testing.M) {
	_ = basemodel.RegisterJsonBaseSchemas()
	os.Exit(m.Run())
}

// Every schema this module registers must have an engine, and vice versa. The two lists are
// maintained by hand in different files; a mismatch is either a resource with no HTTP surface or an
// engine for a schema that does not exist, which fails at boot.
func TestEveryResourceHasAnEngine(t *testing.T) {
	assert.ElementsMatch(t, []string{
		models.ConfigurationSchemaName,
		models.VendorProductPriceSchemaName,
		models.SourcingGroupSchemaName,
		models.AgreementSchemaName,
		models.AgreementLineSchemaName,
		models.PurchaseOrderSchemaName,
		models.PurchaseOrderLineSchemaName,
		models.AuditEventSchemaName,
	}, SchemaNames())
}

// A default field set naming a field the schema lacks produces an engine that fails every listing
// request.
func TestDefaultSearchFieldsExistOnTheirSchemas(t *testing.T) {
	builders := map[string]func() *dmodel.ModelSchemaBuilder{
		models.ConfigurationSchemaName:      models.ConfigurationSchemaBuilder,
		models.VendorProductPriceSchemaName: models.VendorProductPriceSchemaBuilder,
		models.SourcingGroupSchemaName:      models.SourcingGroupSchemaBuilder,
		models.AgreementSchemaName:          models.AgreementSchemaBuilder,
		models.AgreementLineSchemaName:      models.AgreementLineSchemaBuilder,
		models.PurchaseOrderSchemaName:      models.PurchaseOrderSchemaBuilder,
		models.PurchaseOrderLineSchemaName:  models.PurchaseOrderLineSchemaBuilder,
		models.AuditEventSchemaName:         models.AuditEventSchemaBuilder,
	}

	for _, schemaName := range SchemaNames() {
		builder, ok := builders[schemaName]
		require.True(t, ok, "no builder for %s", schemaName)
		schema := builder().Build()

		assert.NotEmptyf(t, schema.DefaultSearchFields(),
			"%s declares no default_search_fields, so its listing returns only primary keys",
			schemaName)
		for _, fieldName := range schema.DefaultSearchFields() {
			_, exists := schema.Field(fieldName)
			assert.True(t, exists, "%s lists default search field %q, which it does not have",
				schemaName, fieldName)
		}
	}
}
