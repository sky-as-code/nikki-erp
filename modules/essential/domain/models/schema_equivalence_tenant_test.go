package models

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The hosted app's apptrait installs a base model carrying a tenant key, so "core.basemodel.base_model"
// adds tenant_id there but not in standalone nikkierp. A JSON model that extends base_model where the
// legacy chain declared only its id passes the single-tenant comparison yet changes the multi-tenant
// migration; repeating the comparison under a tenant base model catches it.
func TestEssentialSchemaJsonMatchesLegacyMultiTenant(t *testing.T) {
	_ = basemodel.RegisterJsonBaseSchemas()
	dmodel.SetBaseModelSchemaBuilder(multiTenantBaseModelForTest())
	t.Cleanup(func() { dmodel.SetBaseModelSchemaBuilder(nil) })

	assertEssentialSchemasMatchLegacy(t)
}

// multiTenantBaseModelForTest mirrors modules/apptrait/dynamicmodel/basemodel.SetBaseModel in the
// coreplatform root, which nikkierp cannot import.
func multiTenantBaseModelForTest() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("_").
		Field(
			basemodel.DefineFieldId("tenant_id").
				Label(model.NewLangJsonRef("fields.tenant_id")).
				UseTypeDefault().
				TenantKey(),
		)
}
