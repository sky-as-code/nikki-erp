package composable

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle/defaults"
)

// CreateDefaults writes a module's `<schema>.json` defaults for one org through svc, resolving
// aliases against those already created. Every row goes through the domain service's Create, so
// ids, etags and the resource's own rules are what a REST create would apply.
func CreateDefaults(
	ctx corectx.Context, svc CrudDomainService, rows []dmodel.DynamicFields, aliases *defaults.Aliases, orgId model.Id,
) error {
	schemaName := svc.Schema().Name()
	err := aliases.CreateAll(rows, orgId, func(row dmodel.DynamicFields) (model.Id, error) {
		return CreateOne(ctx, svc, row)
	})
	return errors.Wrapf(err, "write %s defaults", schemaName)
}

// CreateOne creates one record and returns its id; a client error is an error here, because
// defaults are not user input and a rejected one is a defect.
func CreateOne(ctx corectx.Context, svc CrudDomainService, row dmodel.DynamicFields) (model.Id, error) {
	result, err := svc.Create(ctx, row)
	if err != nil {
		return "", err
	}
	if result.ClientErrors.Count() > 0 {
		return "", result.ClientErrors.ToError()
	}
	if !result.HasData {
		return "", errors.Errorf("create %s returned no record", svc.Schema().Name())
	}
	return ParamId(result.Data), nil
}

// UpdateOne sets fields on the record with this id, reading its current etag first. For wiring
// an edge between defaults after both ends exist.
func UpdateOne(ctx corectx.Context, svc CrudDomainService, id model.Id, fields dmodel.DynamicFields) error {
	found, err := svc.GetById(ctx, dmodel.DynamicFields{basemodel.FieldId: id})
	if err != nil {
		return err
	}
	if found.ClientErrors.Count() > 0 {
		return found.ClientErrors.ToError()
	}
	if !found.HasData {
		return errors.Errorf("update %s: record %s not found", svc.Schema().Name(), id)
	}

	cmd := dmodel.DynamicFields{
		basemodel.FieldId:   id,
		basemodel.FieldEtag: found.Data.Item[basemodel.FieldEtag],
	}
	for key, value := range fields {
		cmd[key] = value
	}
	result, err := svc.Update(ctx, cmd)
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		return errors.Wrapf(result.ClientErrors.ToError(), "update %s %s", svc.Schema().Name(), id)
	}
	return nil
}

// DeleteOrgData deletes everything the org holds in each service's schema, in the order given:
// children before the parents they reference.
func DeleteOrgData(ctx corectx.Context, orgId model.Id, svcs ...CrudDomainService) error {
	for _, svc := range svcs {
		if _, err := svc.DeleteByOrg(ctx, orgId); err != nil {
			return err
		}
	}
	return nil
}

// DeleteOrgDataOf is DeleteOrgData over resources named by schema, children first.
func DeleteOrgDataOf(ctx corectx.Context, orgId model.Id, schemaNames ...string) error {
	svcs, err := ResolveDomainServices(schemaNames...)
	if err != nil {
		return err
	}
	return DeleteOrgData(ctx, orgId, svcs...)
}

// OrgDataCleanup is the org lifecycle of a module that writes no defaults: InitOrg does nothing
// and CleanUpOrg deletes the org's rows of each schema, in the order given (children first), in
// one transaction. A module embeds it in its own OrgLifecycleAppServiceImpl.
type OrgDataCleanup struct {
	schemaNames []string
}

func NewOrgDataCleanup(schemaNames ...string) OrgDataCleanup {
	return OrgDataCleanup{schemaNames: schemaNames}
}

func (this OrgDataCleanup) InitOrg(corectx.Context, model.Id) error {
	return nil
}

func (this OrgDataCleanup) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	if len(this.schemaNames) == 0 {
		return nil
	}
	svcs, err := ResolveDomainServices(this.schemaNames...)
	if err != nil {
		return err
	}
	return InTransaction(ctx, svcs[0].Repository(), func(tranxCtx corectx.Context) error {
		return DeleteOrgData(tranxCtx, orgId, svcs...)
	})
}
