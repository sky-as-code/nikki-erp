package baserepo

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	dEnt "github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
)

func Insert[TDomain any, TDomainPtr dEnt.DynamicModelPtr[TDomain]](
	ctx dEnt.Context, repo dEnt.BaseRepository, domainModel schema.DynamicModelGetter,
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

func Update[TDomain any, TDomainPtr dEnt.DynamicModelPtr[TDomain]](
	ctx dEnt.Context, repo dEnt.BaseRepository, domainModel schema.DynamicModelGetter,
) (TDomainPtr, error) {
	data := domainModel.GetFieldData()
	updated, err := repo.Update(ctx, data)
	if err != nil {
		return nil, err
	}

	var result TDomain
	TDomainPtr(&result).SetFieldData(updated)
	return &result, nil
}

func FindByPk[TDomain any, TDomainPtr dEnt.DynamicModelPtr[TDomain]](
	ctx dEnt.Context, repo dEnt.BaseRepository, domainModel schema.DynamicModelGetter,
) (TDomainPtr, error) {
	found, err := repo.FindByPk(ctx, domainModel.GetFieldData())
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, nil
	}
	var result TDomain
	TDomainPtr(&result).SetFieldData(found)
	return &result, nil
}

func Archive[TDomain any, TDomainPtr dEnt.DynamicModelPtr[TDomain]](
	ctx dEnt.Context, repo dEnt.BaseRepository, domainModel schema.DynamicModelGetter,
) (TDomainPtr, error) {
	archived, err := repo.Archive(ctx, domainModel.GetFieldData())
	if err != nil {
		return nil, err
	}
	if archived == nil {
		return nil, nil
	}
	var result TDomain
	TDomainPtr(&result).SetFieldData(archived)
	return &result, nil
}

func Search[TDomain any, TDomainPtr dEnt.DynamicModelPtr[TDomain]](
	ctx dEnt.Context, repo dEnt.BaseRepository, graph schema.SearchGraph, columns []string,
) ([]TDomainPtr, error) {
	records, err := repo.Search(ctx, graph, columns)
	if err != nil {
		return nil, err
	}
	result := make([]TDomainPtr, len(records))
	for i, record := range records {
		var domain TDomain
		TDomainPtr(&domain).SetFieldData(record)
		result[i] = &domain
	}
	return result, nil
}
