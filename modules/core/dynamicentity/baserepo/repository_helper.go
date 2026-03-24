package baserepo

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	dEnt "github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
)

func Insert[TDomain any, TDomainPtr dEnt.DynamicModelPtr[TDomain]](
	ctx dEnt.Context, repo dEnt.BaseRepository, domainModel schema.DynamicModelGetter,
) (*dEnt.OpResult[TDomain], error) {
	data := domainModel.GetFieldData()
	creation, err := repo.Insert(ctx, data)
	if err != nil {
		return nil, err
	}
	if len(creation.ClientErrors) > 0 {
		return &dEnt.OpResult[TDomain]{ClientErrors: creation.ClientErrors}, nil
	}

	var result TDomain
	TDomainPtr(&result).SetFieldData(creation.Data)
	return &dEnt.OpResult[TDomain]{Data: result, IsEmpty: false}, nil
}

func Update[TDomain any, TDomainPtr dEnt.DynamicModelPtr[TDomain]](
	ctx dEnt.Context, repo dEnt.BaseRepository, domainModel schema.DynamicModelGetter, prevEtag string,
) (*dEnt.OpResult[TDomain], error) {
	data := domainModel.GetFieldData()
	updated, err := repo.Update(ctx, data, prevEtag)
	if err != nil {
		return nil, err
	}
	if len(updated.ClientErrors) > 0 {
		return &dEnt.OpResult[TDomain]{ClientErrors: updated.ClientErrors}, nil
	}

	var result TDomain
	TDomainPtr(&result).SetFieldData(updated.Data)
	return &dEnt.OpResult[TDomain]{Data: result, IsEmpty: false}, nil
}

func FindOne[TDomain any, TDomainPtr dEnt.DynamicModelPtr[TDomain]](
	ctx dEnt.Context, repo dEnt.BaseRepository, findParam dEnt.GetOneParam,
) (*dEnt.OpResult[TDomain], error) {
	found, err := repo.GetOne(ctx, findParam)
	if err != nil {
		return nil, err
	}
	if len(found.ClientErrors) > 0 {
		return &dEnt.OpResult[TDomain]{ClientErrors: found.ClientErrors}, nil
	}
	if found.IsEmpty {
		return &dEnt.OpResult[TDomain]{IsEmpty: true}, nil
	}
	var result TDomain
	TDomainPtr(&result).SetFieldData(found.Data)
	return &dEnt.OpResult[TDomain]{Data: result, IsEmpty: false}, nil
}

func Archive[TDomain any, TDomainPtr dEnt.DynamicModelPtr[TDomain]](
	ctx dEnt.Context, repo dEnt.BaseRepository, domainModel schema.DynamicModelGetter,
) (*dEnt.OpResult[TDomain], error) {
	archived, err := repo.Archive(ctx, domainModel.GetFieldData())
	if err != nil {
		return nil, err
	}
	if len(archived.ClientErrors) > 0 {
		return &dEnt.OpResult[TDomain]{ClientErrors: archived.ClientErrors}, nil
	}
	if archived.IsEmpty {
		return &dEnt.OpResult[TDomain]{IsEmpty: true}, nil
	}
	var result TDomain
	TDomainPtr(&result).SetFieldData(archived.Data)
	return &dEnt.OpResult[TDomain]{Data: result, IsEmpty: false}, nil
}

func Search[TDomain any, TDomainPtr dEnt.DynamicModelPtr[TDomain]](
	ctx dEnt.Context, repo dEnt.BaseRepository, searchParam dEnt.SearchParam,
) (*dEnt.OpResult[dEnt.PagedResult[TDomain]], error) {
	found, err := repo.Search(ctx, searchParam)
	if err != nil {
		return nil, err
	}
	if len(found.ClientErrors) > 0 {
		return &dEnt.OpResult[dEnt.PagedResult[TDomain]]{ClientErrors: found.ClientErrors}, nil
	}
	paged := found.Data
	items := make([]TDomain, len(paged.Items))
	for i, record := range paged.Items {
		var m TDomain
		TDomainPtr(&m).SetFieldData(record)
		items[i] = m
	}
	out := dEnt.PagedResult[TDomain]{
		Items: items,
		Total: paged.Total,
		Page:  paged.Page,
		Size:  paged.Size,
	}
	return &dEnt.OpResult[dEnt.PagedResult[TDomain]]{Data: out, IsEmpty: len(items) == 0}, nil
}
