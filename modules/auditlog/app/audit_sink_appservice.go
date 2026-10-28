package app

import (
	"encoding/json"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/auditlog/domain/models"
	itAction "github.com/sky-as-code/nikki-erp/modules/auditlog/interfaces/action"
	itAuditlog "github.com/sky-as-code/nikki-erp/modules/auditlog/interfaces/auditlog"
	"github.com/sky-as-code/nikki-erp/modules/core/audit"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

const maxDisplayNameLength = 300

func NewAuditSinkAppService(actionSvc itAction.ActionDomainService) itAuditlog.AuditSinkAppService {
	return &AuditSinkAppServiceImpl{actionSvc: actionSvc}
}

// AuditSinkAppServiceImpl implements the core/audit.AuditSink port; apptrait registers it as that
// port. It asserts no permission, by the port's contract, and writes through the domain service
// with the caller's ctx, so every entry joins the transaction of the change it records.
type AuditSinkAppServiceImpl struct {
	actionSvc itAction.ActionDomainService
}

func (this *AuditSinkAppServiceImpl) Record(ctx corectx.Context, entries []audit.Entry) error {
	if len(entries) == 0 || !hasTenantScope(ctx) {
		return nil
	}
	// A meta minted here is stored on ctx, so the other entries this ctx writes share it.
	meta := corectx.EnsureRequestMeta(ctx, corectx.RequestSourceSystem)
	actor := actorOf(ctx)
	for _, entry := range entries {
		row, err := buildActionRow(entry, actor, meta)
		if err != nil {
			return err
		}
		result, err := this.actionSvc.Create(ctx, row)
		if err != nil {
			return errors.Wrapf(err, "record audit entry %s %s", entry.Action, entry.ResourceCode)
		}
		// An entry is built here, not by a client, so a rejected one is a defect and must fail the
		// action it was meant to record rather than let it commit unrecorded.
		if result.ClientErrors.Count() > 0 {
			return errors.Wrapf(result.ClientErrors.ToError(),
				"audit entry %s %s was rejected", entry.Action, entry.ResourceCode)
		}
	}
	return nil
}

// hasTenantScope is false only in the hosted app, where entries carry a tenant key, for a write made
// with no tenant in scope: a global, platform-level record such as the module metadata synced at
// start-up. No tenant owns such a write, so its entry has no tenant to be stored under.
func hasTenantScope(ctx corectx.Context) bool {
	return dmodel.MustGetSchema(models.ActionSchemaName).TenantKey() == "" || ctx.GetDomainConstraints() != nil
}

type actorSnapshot struct {
	kind        string
	userId      *model.Id
	displayName *string
}

// actorOf reads who is acting. A job or consumer may act with no user, in which case the entry
// records the principal's kind and name only; with no principal at all it is the system.
func actorOf(ctx corectx.Context) actorSnapshot {
	perms := ctx.GetPermissions()
	actor := actorSnapshot{kind: string(perms.Principal.Kind)}
	if actor.kind == "" {
		actor.kind = string(corectx.PrincipalKindSystem)
	}
	if perms.UserId != "" {
		userId := perms.UserId
		actor.userId = &userId
	}
	name := perms.Principal.DisplayName
	if user := ctx.GetUser(); user != nil {
		if display, ok := user["display_name"].(string); ok && display != "" {
			name = display
		}
	}
	if len(name) > maxDisplayNameLength {
		name = name[:maxDisplayNameLength]
	}
	if name != "" {
		actor.displayName = &name
	}
	return actor
}

func buildActionRow(entry audit.Entry, actor actorSnapshot, meta corectx.RequestMeta) (dmodel.DynamicFields, error) {
	before, err := plainJsonMap(entry.Before)
	if err != nil {
		return nil, err
	}
	after, err := plainJsonMap(entry.After)
	if err != nil {
		return nil, err
	}
	row := dmodel.DynamicFields{
		models.ActionFieldAction:       entry.Action,
		models.ActionFieldResourceCode: entry.ResourceCode,
		models.ActionFieldActorType:    actor.kind,
		models.ActionFieldBefore:       before,
		models.ActionFieldAfter:        after,
		models.ActionFieldSource:       string(sourceOrSystem(meta.Source)),
	}
	setIfPresent(row, models.ActionFieldResourceId, entry.ResourceId)
	setIfPresent(row, models.ActionFieldActorDisplayName, actor.displayName)
	if entry.OrgId != nil {
		row[models.ActionFieldOrgId] = string(*entry.OrgId)
	}
	if actor.userId != nil {
		row[models.ActionFieldActorId] = string(*actor.userId)
	}
	if meta.Id != "" {
		row[models.ActionFieldRequestId] = meta.Id
	}
	return row, nil
}

func setIfPresent(row dmodel.DynamicFields, field string, value *string) {
	if value != nil && *value != "" {
		row[field] = *value
	}
}

func sourceOrSystem(source corectx.RequestSource) corectx.RequestSource {
	switch source {
	case corectx.RequestSourceHttp, corectx.RequestSourceJob, corectx.RequestSourceEvent:
		return source
	}
	return corectx.RequestSourceSystem
}

// plainJsonMap reduces the values to what they serialize as (a datetime to its string, a decimal
// to its text), so the jsonmap column stores exactly what a reader will get back.
func plainJsonMap(fields dmodel.DynamicFields) (map[string]any, error) {
	out := map[string]any{}
	if len(fields) == 0 {
		return out, nil
	}
	raw, err := json.Marshal(fields)
	if err != nil {
		return nil, errors.Wrap(err, "plainJsonMap.Marshal")
	}
	return out, errors.Wrap(json.Unmarshal(raw, &out), "plainJsonMap.Unmarshal")
}
