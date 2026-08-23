package models

import (
	"encoding/json"
	"fmt"
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func TestZZDumpIamFields(t *testing.T) {
	_ = basemodel.RegisterJsonBaseSchemas()
	out := map[string]any{}
	for name, b := range map[string]*dmodel.ModelSchemaBuilder{
		"iam_action":        ActionSchemaBuilder(),
		"iam_entitlement":   EntitlementSchemaBuilder(),
		"iam_grant_request": RoleRequestSchemaBuilder(),
		"iam_org":           OrganizationSchemaBuilder(),
		"iam_orgunit":       OrganizationalUnitSchemaBuilder(),
		"iam_resource":      ResourceSchemaBuilder(),
		"iam_user":          UserSchemaBuilder(),
	} {
		s := b.Build()
		f := map[string]any{}
		for n, fl := range s.Fields() {
			f[n] = map[string]any{"fk": fl.IsForeignKey(), "edge": fl.IsEdgeModel(), "type": string(fl.ColumnType())}
		}
		out[name] = map[string]any{"label": s.RecordLabelField(), "fields": f}
	}
	b, _ := json.Marshal(out)
	fmt.Println("IAMDUMP" + string(b))
}
