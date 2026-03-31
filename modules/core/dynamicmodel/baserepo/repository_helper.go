package baserepo

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func GetOne[TDomain any, TDomainPtr dyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context, dynamicRepo dyn.BaseRepository, param dyn.RepoGetOneParam,
) (*dyn.OpResult[TDomain], error) {
	found, err := dynamicRepo.GetOne(ctx, param)
	if err != nil {
		return nil, err
	}
	if len(found.ClientErrors) > 0 {
		return &dyn.OpResult[TDomain]{ClientErrors: found.ClientErrors}, nil
	}
	if !found.HasData {
		return &dyn.OpResult[TDomain]{HasData: false}, nil
	}
	var result TDomain
	TDomainPtr(&result).SetFieldData(found.Data)
	return &dyn.OpResult[TDomain]{Data: result, HasData: true}, nil
}

func DeleteOne(
	ctx corectx.Context, dynamicRepo dyn.BaseRepository, keys dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	delResult, err := dynamicRepo.DeleteOne(ctx, keys)
	if err != nil {
		return nil, err
	}
	if delResult.ClientErrors.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: delResult.ClientErrors}, nil
	}
	if !delResult.HasData {
		return &dyn.OpResult[dyn.MutateResultData]{HasData: false}, nil
	}
	return &dyn.OpResult[dyn.MutateResultData]{
		Data: dyn.MutateResultData{
			AffectedCount: delResult.Data,
			AffectedAt:    model.NewModelDateTime(),
		},
		HasData: true,
	}, nil
}

func Exists(
	ctx corectx.Context, dynamicRepo dyn.BaseRepository, keys []dmodel.DynamicFields,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	return dynamicRepo.Exists(ctx, keys)
}

func Insert(
	ctx corectx.Context, dynamicRepo dyn.BaseRepository, domainModel dmodel.DynamicModelGetter,
) (*dyn.OpResult[int], error) {
	data := domainModel.GetFieldData()
	creation, err := dynamicRepo.Insert(ctx, data)
	if err != nil {
		return nil, err
	}
	if len(creation.ClientErrors) > 0 {
		return &dyn.OpResult[int]{ClientErrors: creation.ClientErrors}, nil
	}

	return &dyn.OpResult[int]{Data: creation.Data, HasData: true}, nil
}

func Search[TDomain any, TDomainPtr dyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context, dynamicRepo dyn.BaseRepository, searchParam dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[TDomain]], error) {
	found, err := dynamicRepo.Search(ctx, searchParam)
	if err != nil {
		return nil, err
	}
	if len(found.ClientErrors) > 0 {
		return &dyn.OpResult[dyn.PagedResultData[TDomain]]{ClientErrors: found.ClientErrors}, nil
	}
	paged := found.Data
	items := make([]TDomain, len(paged.Items))
	for i, record := range paged.Items {
		var m TDomain
		TDomainPtr(&m).SetFieldData(record)
		items[i] = m
	}
	out := dyn.PagedResultData[TDomain]{
		Items: items,
		Total: paged.Total,
		Page:  paged.Page,
		Size:  paged.Size,
	}

	return &dyn.OpResult[dyn.PagedResultData[TDomain]]{
		Data:    out,
		HasData: len(items) != 0,
	}, nil
}

func Update(
	ctx corectx.Context, dynamicRepo dyn.BaseRepository, data dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	updatedRes, err := dynamicRepo.Update(ctx, data)
	if err != nil {
		return nil, err
	}
	if len(updatedRes.ClientErrors) > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: updatedRes.ClientErrors}, nil
	}
	updatedAt, etag := readUpdatedAtAndEtagFromFields(updatedRes.Data)
	return &dyn.OpResult[dyn.MutateResultData]{
		Data: dyn.MutateResultData{
			AffectedCount: 1,
			AffectedAt:    updatedAt,
			Etag:          etag,
		},
		HasData: true,
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
