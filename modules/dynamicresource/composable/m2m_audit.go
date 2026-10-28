package composable

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/audit"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// sharedAuditSink serves the writes made outside any onion: the m2m actions and RecordAudit.
var sharedAuditSink AuditRecorder = audit.NewLazyAuditSink()

// recordM2mChange writes a link entry for the added peers and an unlink entry for the removed
// ones, inside the manage call's transaction, on the source record and on every peer: the history
// of either party then reads "roles: + A, B". On a peer the field is its back edge, the value the
// source id.
func recordM2mChange(ctx corectx.Context, repo CrudRepository, destSchema string, change M2mChange) error {
	schema := repo.Schema()
	if schema == nil || !sharedAuditSink.Enabled() {
		return nil
	}
	orgId, err := m2mSourceOrg(ctx, repo, schema, change.SrcId)
	if err != nil {
		return err
	}
	entries := []audit.Entry{}
	if IsAuditedSchema(schema.Name()) {
		entries = linkEntries(schema.Name(), string(change.SrcId), orgId,
			m2mEdgeName(schema, destSchema), idStrings(change.Added), idStrings(change.Removed))
	}
	entries = append(entries, peerLinkEntries(schema, m2mDestSchema(schema, destSchema), change, orgId)...)
	if len(entries) == 0 {
		return nil
	}
	return errors.Wrapf(sharedAuditSink.Record(ctx, entries), "record m2m change of %s", schema.Name())
}

// RecordLinkAudit is for a custom write that links or unlinks records outside ManageM2m, such as
// the entitlement delta of a role. Call it inside the write's transaction; field names what was
// linked, the values are how the history shows them.
func RecordLinkAudit(
	ctx corectx.Context, schema *dmodel.ModelSchema, srcId model.Id, orgId *model.Id,
	field string, added, removed []string,
) error {
	if schema == nil || !IsAuditedSchema(schema.Name()) || !sharedAuditSink.Enabled() {
		return nil
	}
	entries := linkEntries(schema.Name(), string(srcId), orgId, field, added, removed)
	if len(entries) == 0 {
		return nil
	}
	return errors.Wrapf(sharedAuditSink.Record(ctx, entries), "record link change of %s", schema.Name())
}

func linkEntries(
	resourceCode string, resourceId string, orgId *model.Id, field string, added, removed []string,
) []audit.Entry {
	entries := []audit.Entry{}
	if len(added) > 0 {
		entries = append(entries, audit.Entry{
			Action: audit.ActionLink, ResourceCode: resourceCode, ResourceId: &resourceId, OrgId: orgId,
			Before: dmodel.DynamicFields{}, After: dmodel.DynamicFields{field: added},
		})
	}
	if len(removed) > 0 {
		entries = append(entries, audit.Entry{
			Action: audit.ActionUnlink, ResourceCode: resourceCode, ResourceId: &resourceId, OrgId: orgId,
			Before: dmodel.DynamicFields{field: removed}, After: dmodel.DynamicFields{},
		})
	}
	return entries
}

// peerLinkEntries mirrors the change on each peer. The peer takes the source's org rather than
// its own, which would cost a read per peer; the history is looked up by record, not by org.
func peerLinkEntries(
	src *dmodel.ModelSchema, dest *dmodel.ModelSchema, change M2mChange, srcOrgId *model.Id,
) []audit.Entry {
	if dest == nil || !IsAuditedSchema(dest.Name()) {
		return nil
	}
	var orgId *model.Id
	if _, hasOrg := dest.Field(basemodel.FieldOrgId); hasOrg {
		orgId = srcOrgId
	}
	field := m2mEdgeName(dest, src.Name())
	srcIds := []string{string(change.SrcId)}
	entries := []audit.Entry{}
	for _, peerId := range change.Added {
		entries = append(entries, linkEntries(dest.Name(), string(peerId), orgId, field, srcIds, nil)...)
	}
	for _, peerId := range change.Removed {
		entries = append(entries, linkEntries(dest.Name(), string(peerId), orgId, field, nil, srcIds)...)
	}
	return entries
}

func m2mDestSchema(schema *dmodel.ModelSchema, destSchema string) *dmodel.ModelSchema {
	if link, ok := schema.M2mPeerLinkForDest(destSchema); ok && link.DestSchema != nil {
		return link.DestSchema
	}
	return dmodel.GetSchema(destSchema)
}

func idStrings(ids []model.Id) []string {
	out := make([]string, len(ids))
	for i := range ids {
		out[i] = string(ids[i])
	}
	return out
}

func m2mEdgeName(schema *dmodel.ModelSchema, destSchema string) string {
	if link, ok := schema.M2mPeerLinkForDest(destSchema); ok && link.Edge != "" {
		return link.Edge
	}
	return destSchema
}

// m2mSourceOrg reads the source record's org, which the manage command does not carry.
func m2mSourceOrg(
	ctx corectx.Context, repo CrudRepository, schema *dmodel.ModelSchema, srcId model.Id,
) (*model.Id, error) {
	if _, hasOrg := schema.Field(basemodel.FieldOrgId); !hasOrg {
		return nil, nil
	}
	found, err := repo.FindByKeys(ctx, dmodel.DynamicFields{basemodel.FieldId: srcId})
	if err != nil {
		return nil, errors.Wrap(err, "m2mSourceOrg")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return auditOrgOf(found.Data), nil
}
