package baserepo

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	dEnt "github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
)

func Insert[TDomain any, TDomainPtr dEnt.DynamicModelPtr[TDomain]](
	ctx dEnt.Context, repo dEnt.DbRepository, domainModel schema.DynamicModelGetter,
) (TDomainPtr, error) {
	data := domainModel.GetFieldData()
	creation, err := repo.Insert(ctx, data)
	if err != nil {
		return nil, err
	}

	var result TDomain
	TDomainPtr(&result).SetFieldData(creation)
	return &result, nil
}
