package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/audit"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

func itAuditEntry() audit.Entry {
	id := "r1"
	return audit.Entry{
		Action: audit.ActionCreate, ResourceCode: "purchase_order", ResourceId: &id,
		After: dmodel.DynamicFields{"name": "n"},
	}
}

func metaWithoutSource() corectx.RequestMeta {
	return corectx.RequestMeta{Id: "req_1"}
}
