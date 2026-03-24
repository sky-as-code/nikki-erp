package baserepo

import (
	crud "github.com/sky-as-code/nikki-erp/common/crud"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coredyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

func Insert[TDomain any, TDomainPtr coredyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context, repo coredyn.BaseRepository, domainModel dmodel.DynamicModelGetter,
) (*crud.OpResult[TDomain], error) {
	data := domainModel.GetFieldData()
	creation, err := repo.Insert(ctx, data)
	if err != nil {
		return nil, err
	}
	if len(creation.ClientErrors) > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: creation.ClientErrors}, nil
	}

	var result TDomain
	TDomainPtr(&result).SetFieldData(creation.Data)
	return &crud.OpResult[TDomain]{Data: result, IsEmpty: false}, nil
}

func Update[TDomain any, TDomainPtr coredyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context, repo coredyn.BaseRepository, domainModel dmodel.DynamicModelGetter, prevEtag string,
) (*crud.OpResult[TDomain], error) {
	data := domainModel.GetFieldData()
	updated, err := repo.Update(ctx, data, prevEtag)
	if err != nil {
		return nil, err
	}
	if len(updated.ClientErrors) > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: updated.ClientErrors}, nil
	}

	var result TDomain
	TDomainPtr(&result).SetFieldData(updated.Data)
	return &crud.OpResult[TDomain]{Data: result, IsEmpty: false}, nil
}

func FindOne[TDomain any, TDomainPtr coredyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context, repo coredyn.BaseRepository, findParam coredyn.GetOneParam,
) (*crud.OpResult[TDomain], error) {
	found, err := repo.GetOne(ctx, findParam)
	if err != nil {
		return nil, err
	}
	if len(found.ClientErrors) > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: found.ClientErrors}, nil
	}
	if found.IsEmpty {
		return &crud.OpResult[TDomain]{IsEmpty: true}, nil
	}
	var result TDomain
	TDomainPtr(&result).SetFieldData(found.Data)
	return &crud.OpResult[TDomain]{Data: result, IsEmpty: false}, nil
}

func Archive[TDomain any, TDomainPtr coredyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context, repo coredyn.BaseRepository, domainModel dmodel.DynamicModelGetter,
) (*crud.OpResult[TDomain], error) {
	archived, err := repo.Archive(ctx, domainModel.GetFieldData())
	if err != nil {
		return nil, err
	}
	if len(archived.ClientErrors) > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: archived.ClientErrors}, nil
	}
	if archived.IsEmpty {
		return &crud.OpResult[TDomain]{IsEmpty: true}, nil
	}
	var result TDomain
	TDomainPtr(&result).SetFieldData(archived.Data)
	return &crud.OpResult[TDomain]{Data: result, IsEmpty: false}, nil
}

func Search[TDomain any, TDomainPtr coredyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context, repo coredyn.BaseRepository, searchParam coredyn.SearchParam,
) (*crud.OpResult[crud.PagedResult[TDomain]], error) {
	found, err := repo.Search(ctx, searchParam)
	if err != nil {
		return nil, err
	}
	if len(found.ClientErrors) > 0 {
		return &crud.OpResult[crud.PagedResult[TDomain]]{ClientErrors: found.ClientErrors}, nil
	}
	paged := found.Data
	items := make([]TDomain, len(paged.Items))
	for i, record := range paged.Items {
		var m TDomain
		TDomainPtr(&m).SetFieldData(record)
		items[i] = m
	}
	out := crud.PagedResult[TDomain]{
		Items: items,
		Total: paged.Total,
		Page:  paged.Page,
		Size:  paged.Size,
	}
	return &crud.OpResult[crud.PagedResult[TDomain]]{Data: out, IsEmpty: len(items) == 0}, nil
}
