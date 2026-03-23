package drive_file_share_service_impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

func (this *DriveFileShareServiceImpl) GetDriveFileShareById(ctx crud.Context, query it.GetDriveFileShareByIdQuery) (
	*it.GetDriveFileShareByIdResult, error) {
	return crud.GetOne(ctx, crud.GetOneParam[*domain.DriveFileShare, it.GetDriveFileShareByIdQuery, it.GetDriveFileShareByIdResult]{
		Action: "get drive file share by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q it.GetDriveFileShareByIdQuery, vErrs *ft.ValidationErrors) (*domain.DriveFileShare, error) {
			share, err := this.driveFileShareRepo.FindById(ctx, q.DriveFileShareId)
			if err != nil {
				return nil, err
			}
			if share == nil {
				vErrs.AppendNotFound("driveFileShareId", "drive file share")
			}
			if share != nil {
				if err := this.enrichDriveFileSharesWithUsers(ctx, []*domain.DriveFileShare{share}); err != nil {
					return nil, err
				}
			}
			return share, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetDriveFileShareByIdResult {
			return &it.GetDriveFileShareByIdResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(d *domain.DriveFileShare) *it.GetDriveFileShareByIdResult {
			return &it.GetDriveFileShareByIdResult{HasData: true, Data: d}
		},
	})
}

// Get by File ID
func (this *DriveFileShareServiceImpl) GetDriveFileShareByFileId(ctx crud.Context, query it.GetDriveFileShareByFileIdQuery) (
	*it.GetDriveFileShareByFileIdResult, error) {
	return crud.Search(ctx, crud.SearchParam[*domain.DriveFileShare, it.GetDriveFileShareByFileIdQuery, it.GetDriveFileShareByFileIdResult]{
		Action: "get drive files by parent",
		Query:  query,
		SetQueryDefaults: func(q *it.GetDriveFileShareByFileIdQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: this.driveFileRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, q it.GetDriveFileShareByFileIdQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[*domain.DriveFileShare], error) {
			paged, err := this.driveFileShareRepo.ListByFileRef(ctx, it.ListByFileRefParam{
				FileRef: q.DriveFileId,
				SearchParam: it.SearchParam{
					Predicate: predicate,
					Order:     order,
					Page:      *q.Page,
					Size:      *q.Size,
				},
			})
			if err != nil {
				return nil, err
			}
			if paged != nil {
				if err := this.enrichDriveFileSharesWithUsers(ctx, paged.Items); err != nil {
					return nil, err
				}
			}
			return paged, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetDriveFileShareByFileIdResult {
			return &it.GetDriveFileShareByFileIdResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(paged *crud.PagedResult[*domain.DriveFileShare]) *it.GetDriveFileShareByFileIdResult {
			return &it.GetDriveFileShareByFileIdResult{Data: paged, HasData: true}
		},
	})
}

// Get by User

func (this *DriveFileShareServiceImpl) GetDriveFileShareByUser(ctx crud.Context, query it.GetDriveFileShareByUserQuery) (
	*it.GetDriveFileShareByUserResult, error) {
	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetDriveFileShareByUserResult{ClientError: vErrs.ToClientError()}, nil
	}

	items, err := this.driveFileShareRepo.ListByUserRef(ctx, query.UserId)
	if err != nil {
		return nil, err
	}
	if err := this.enrichDriveFileSharesWithUsers(ctx, items); err != nil {
		return nil, err
	}

	return &it.GetDriveFileShareByUserResult{
		HasData: items != nil,
		Data: &it.GetDriveFileShareByUserResultData{
			Items: items,
			Total: len(items),
		},
	}, nil
}

// Search

func (this *DriveFileShareServiceImpl) SearchDriveFileShare(ctx crud.Context, query it.SearchDriveFileShareQuery) (
	*it.SearchDriveFileShareResult, error) {
	return crud.Search(ctx, crud.SearchParam[*domain.DriveFileShare, it.SearchDriveFileShareQuery, it.SearchDriveFileShareResult]{
		Action: "search drive file shares",
		Query:  query,
		SetQueryDefaults: func(q *it.SearchDriveFileShareQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: this.driveFileShareRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, q it.SearchDriveFileShareQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[*domain.DriveFileShare], error) {
			paged, err := this.driveFileShareRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *q.Page,
				Size:      *q.Size,
			})
			if err != nil {
				return nil, err
			}
			if paged != nil {
				if err := this.enrichDriveFileSharesWithUsers(ctx, paged.Items); err != nil {
					return nil, err
				}
			}
			return paged, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchDriveFileShareResult {
			return &it.SearchDriveFileShareResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(paged *crud.PagedResult[*domain.DriveFileShare]) *it.SearchDriveFileShareResult {
			return &it.SearchDriveFileShareResult{Data: paged, HasData: paged.Items != nil}
		},
	})
}
