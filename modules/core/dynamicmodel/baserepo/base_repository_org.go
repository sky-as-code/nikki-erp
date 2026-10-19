package baserepo

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// DeleteByOrg deletes every row of this schema that belongs to orgId, within the context's
// tenant. It serves org cleanup only: a foreign key that refuses the delete is returned as an
// error, not a client error, because it means a module cleaned up in the wrong order.
func (this *BaseDynamicRepositoryImpl) DeleteByOrg(ctx corectx.Context, orgId model.Id) (*dyn.OpResult[int], error) {
	if orgId == "" {
		return nil, errors.Errorf("DeleteByOrg on %s: no org id", this.schema.Name())
	}
	if _, found := this.schema.Column(basemodel.FieldOrgId); !found {
		return nil, errors.Errorf("DeleteByOrg on %s: the schema has no %s column", this.schema.Name(), basemodel.FieldOrgId)
	}

	filter := injectTenantFilter(ctx, this.schema, dmodel.DynamicFields{basemodel.FieldOrgId: string(orgId)})
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlDeleteEqual(this.schema, filter)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return nil, errors.Wrapf(qbClientErrs.ToError(), "DeleteByOrg on %s", this.schema.Name())
	}

	this.logQuery(*sqlQuery)
	result, err := this.ExtractClient(ctx).Exec(ctx, *sqlQuery)
	if err != nil {
		return nil, errors.Wrapf(err, "DeleteByOrg on %s", this.schema.Name())
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[int]{Data: int(affected), HasData: affected != 0}, nil
}
