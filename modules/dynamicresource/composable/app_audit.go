package composable

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/audit"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// Read logging is opt-in (AuditReads): only a successful read is recorded, after the
// authorization and the masking, and a failure to record it fails the read, as a mutation's does.

func (this *DefaultApplicationServiceImpl) auditGet(
	ctx corectx.Context, orgId *model.Id, query GetByIdQuery, result *GetOneResult, readErr error,
) error {
	if readErr != nil || !this.readAuditActive() || result == nil || !result.HasData || result.ClientErrors.Count() > 0 {
		return readErr
	}
	id := readString(query, basemodel.FieldId)
	entry := audit.Entry{
		Action: audit.ActionGet, ResourceCode: this.ResourceCode(), ResourceId: &id,
		OrgId: orgId, Before: dmodel.DynamicFields{}, After: dmodel.DynamicFields{},
	}
	return errors.Wrap(this.readAudit.Record(ctx, []audit.Entry{entry}), "audit "+this.ResourceCode()+" get")
}

func (this *DefaultApplicationServiceImpl) auditSearch(
	ctx corectx.Context, orgId *model.Id, query SearchQuery, result *SearchResult, readErr error,
) error {
	if readErr != nil || !this.readAuditActive() || result == nil || result.ClientErrors.Count() > 0 {
		return readErr
	}
	asked := dmodel.DynamicFields{}
	for _, field := range []string{fieldNameGraph, fieldNameFields} {
		if value, ok := query[field]; ok && value != nil {
			asked[field] = value
		}
	}
	entry := audit.Entry{
		Action: audit.ActionSearch, ResourceCode: this.ResourceCode(),
		OrgId: orgId, Before: dmodel.DynamicFields{}, After: asked,
	}
	return errors.Wrap(this.readAudit.Record(ctx, []audit.Entry{entry}), "audit "+this.ResourceCode()+" search")
}

func (this *DefaultApplicationServiceImpl) readAuditActive() bool {
	return this.readAudit != nil && this.readAudit.Enabled()
}
