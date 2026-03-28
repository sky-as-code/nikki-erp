package baserepo

import (
	crud "github.com/sky-as-code/nikki-erp/common/crud"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coredyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
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

func GetOne[TDomain any, TDomainPtr coredyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context, repo coredyn.BaseRepository, findParam coredyn.RepoGetOneParam,
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

func Search[TDomain any, TDomainPtr coredyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context, repo coredyn.BaseRepository, searchParam coredyn.RepoSearchParam,
) (*crud.OpResult[crud.PagedResultData[TDomain]], error) {
	found, err := repo.Search(ctx, searchParam)
	if err != nil {
		return nil, err
	}
	if len(found.ClientErrors) > 0 {
		return &crud.OpResult[crud.PagedResultData[TDomain]]{ClientErrors: found.ClientErrors}, nil
	}
	paged := found.Data
	items := make([]TDomain, len(paged.Items))
	for i, record := range paged.Items {
		var m TDomain
		TDomainPtr(&m).SetFieldData(record)
		items[i] = m
	}
	out := crud.PagedResultData[TDomain]{
		Items: items,
		Total: paged.Total,
		Page:  paged.Page,
		Size:  paged.Size,
	}
	return &crud.OpResult[crud.PagedResultData[TDomain]]{Data: out, IsEmpty: len(items) == 0}, nil
}

func Update[TDomain any, TDomainPtr coredyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context, repo coredyn.BaseRepository, domainModel dmodel.DynamicModelGetter,
) (*crud.OpResult[TDomain], error) {
	data := domainModel.GetFieldData()
	updated, err := repo.Update(ctx, data)
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

func DeleteOne(
	ctx corectx.Context,
	repo coredyn.BaseRepository,
	keys dmodel.DynamicFields,
) (*crud.OpResult[crud.MutateResultData], error) {
	delResult, err := repo.DeleteOne(ctx, keys)
	if err != nil {
		return nil, err
	}
	if delResult.ClientErrors.Count() > 0 {
		return &crud.OpResult[crud.MutateResultData]{ClientErrors: delResult.ClientErrors}, nil
	}
	if delResult.IsEmpty {
		return &crud.OpResult[crud.MutateResultData]{IsEmpty: true}, nil
	}
	return &crud.OpResult[crud.MutateResultData]{
		Data: crud.MutateResultData{
			AffectedCount: delResult.Data,
			AffectedAt:    model.NewModelDateTime(),
		},
	}, nil
}

func UpdateMutate(
	ctx corectx.Context,
	repo coredyn.BaseRepository,
	data dmodel.DynamicFields,
) (*crud.OpResult[crud.MutateResultData], error) {
	updatedRes, err := repo.Update(ctx, data)
	if err != nil {
		return nil, err
	}
	if len(updatedRes.ClientErrors) > 0 {
		return &crud.OpResult[crud.MutateResultData]{ClientErrors: updatedRes.ClientErrors}, nil
	}
	updatedAt, etag := readUpdatedAtAndEtagFromFields(updatedRes.Data)
	return &crud.OpResult[crud.MutateResultData]{
		Data: crud.MutateResultData{
			AffectedCount: 1,
			AffectedAt:    updatedAt,
			Etag:          etag,
		},
	}, nil
}

func readUpdatedAtAndEtagFromFields(data dmodel.DynamicFields) (updatedAt model.ModelDateTime, etag model.Etag) {
	upt, ok := data[basemodel.FieldUpdatedAt]
	if ok {
		updatedAt = upt.(model.ModelDateTime)
	}
	et, ok := data[basemodel.FieldEtag]
	if ok {
		etag = et.(string)
	}
	return
}
