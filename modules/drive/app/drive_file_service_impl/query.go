package drive_file_service_impl

import (
	"io"

	"github.com/sky-as-code/nikki-erp/common/fault"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/external/file_storage"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

func (this *DriveFileServiceImpl) GetDriveFileById(ctx crud.Context, query it.GetDriveFileByIdQuery) (result *it.GetDriveFileByIdResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get drive file by id"); e != nil {
			err = e
		}
	}()
	result, err = crud.GetOne(ctx, crud.GetOneParam[*domain.DriveFile, it.GetDriveFileByIdQuery, it.GetDriveFileByIdResult]{
		Action: "get drive file by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q it.GetDriveFileByIdQuery, vErrs *ft.ValidationErrors) (*domain.DriveFile, error) {
			driveFile, err := this.driveFileRepo.FindById(ctx, q.DriveFileId)
			ft.PanicOnErr(err)
			if driveFile == nil {
				vErrs.AppendNotFound("driveFileId", "drive file")
				return nil, nil
			}

			perm, err := this.resolvePermission(ctx, driveFile, q.UserId)
			ft.PanicOnErr(err)
			if perm < enum.DriveFileSharePermView {
				vErrs.AppendNotAllowed("driveFileId", "drive file")
				return nil, nil
			}
			return driveFile, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetDriveFileByIdResult {
			return &it.GetDriveFileByIdResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(d *domain.DriveFile) *it.GetDriveFileByIdResult {
			return &it.GetDriveFileByIdResult{HasData: true, Data: d}
		},
	})
	return result, err
}

func (this *DriveFileServiceImpl) DownloadDriveFile(ctx crud.Context, query it.GetDriveFileByIdQuery) (d *domain.DriveFile, rc io.ReadCloser, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "download drive file"); e != nil {
			err = e
		}
	}()
	driveFile, err := this.driveFileRepo.FindById(ctx, query.DriveFileId)
	ft.PanicOnErr(err)

	if driveFile == nil {
		return nil, nil, nil
	}

	perm, err := this.resolvePermission(ctx, driveFile, query.UserId)
	ft.PanicOnErr(err)
	if perm < enum.DriveFileSharePermView {
		return nil, nil, &fault.ClientError{
			Code:    "forbidden",
			Details: "drive file is not allowed",
		}
	}

	ioReader, err := this.storageAdapter.DownloadBucket(ctx.InnerContext(), file_storage.BucketDrive, driveFile.StorageKey)
	ft.PanicOnErr(err)

	return driveFile, ioReader, nil
}

// Get by parent

func (this *DriveFileServiceImpl) GetDriveFileByParent(ctx crud.Context, query it.GetDriveFileByParentQuery) (result *it.GetDriveFileByParentResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get drive file by parent"); e != nil {
			err = e
		}
	}()
	query.SetDefaults()
	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetDriveFileByParentResult{ClientError: vErrs.ToClientError()}, nil
	}

	// Assert parent exists when listing children of a specific folder (skip for root: empty parent id)
	if query.FileParentId != "" {
		vErrsAssert := ft.NewValidationErrors()
		parentModel := &domain.DriveFile{ModelBase: model.ModelBase{Id: &query.FileParentId}}

		parent, err := this.assertDriveFileExists(ctx, parentModel, &vErrsAssert)
		ft.PanicOnErr(err)

		if parent == nil {
			return &it.GetDriveFileByParentResult{ClientError: vErrsAssert.ToClientError()}, nil
		}

		perm, err := this.resolvePermission(ctx, parent, query.UserId)
		ft.PanicOnErr(err)
		if perm < enum.DriveFileSharePermView {
			return &it.GetDriveFileByParentResult{
				ClientError: &fault.ClientError{
					Code:    "forbidden",
					Details: "drive file parent is not allowed",
				},
			}, nil
		}
	}

	if query.FileParentId == "" {
		return crud.Search(ctx, crud.SearchParam[*domain.DriveFile, it.GetDriveFileByParentQuery, it.GetDriveFileByParentResult]{
			Action: "get root drive files",
			Query:  query,
			SetQueryDefaults: func(q *it.GetDriveFileByParentQuery) {
				q.SetDefaults()
			},
			ParseSearchGraph: this.driveFileRepo.ParseSearchGraph,
			RepoSearch: func(ctx crud.Context, q it.GetDriveFileByParentQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[*domain.DriveFile], error) {
				return this.driveFileRepo.GetRootFileByUser(ctx, q.UserId, it.SearchParam{
					Predicate: predicate,
					Order:     order,
					Page:      *q.Page,
					Size:      *q.Size,
				})
			},
			ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetDriveFileByParentResult {
				return &it.GetDriveFileByParentResult{ClientError: vErrs.ToClientError()}
			},
			ToSuccessResult: func(paged *crud.PagedResult[*domain.DriveFile]) *it.GetDriveFileByParentResult {
				return &it.GetDriveFileByParentResult{Data: paged, HasData: true}
			},
		})
	}

	return crud.Search(ctx, crud.SearchParam[*domain.DriveFile, it.GetDriveFileByParentQuery, it.GetDriveFileByParentResult]{
		Action: "get drive files by parent",
		Query:  query,
		SetQueryDefaults: func(q *it.GetDriveFileByParentQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: this.driveFileRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, q it.GetDriveFileByParentQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[*domain.DriveFile], error) {
			return this.driveFileRepo.SearchByParent(ctx, it.SearchByParentParam{
				ParentFileId: q.FileParentId,
				Predicate:    predicate,
				Order:        order,
				Page:         *q.Page,
				Size:         *q.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetDriveFileByParentResult {
			return &it.GetDriveFileByParentResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(paged *crud.PagedResult[*domain.DriveFile]) *it.GetDriveFileByParentResult {
			return &it.GetDriveFileByParentResult{Data: paged, HasData: true}
		},
	})
}

func (this *DriveFileServiceImpl) SearchDriveFile(ctx crud.Context, query it.SearchDriveFileQuery) (result *it.SearchDriveFileResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "search drive file"); e != nil {
			err = e
		}
	}()
	result, err = crud.Search(ctx, crud.SearchParam[*domain.DriveFile, it.SearchDriveFileQuery, it.SearchDriveFileResult]{
		Action: "search drive files",
		Query:  query,
		SetQueryDefaults: func(q *it.SearchDriveFileQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: this.driveFileRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, q it.SearchDriveFileQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[*domain.DriveFile], error) {
			return this.driveFileRepo.SearchAccessible(ctx, q.UserId, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *q.Page,
				Size:      *q.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchDriveFileResult {
			return &it.SearchDriveFileResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(paged *crud.PagedResult[*domain.DriveFile]) *it.SearchDriveFileResult {
			return &it.SearchDriveFileResult{Data: paged, HasData: paged.Items != nil}
		},
	})
	return result, err
}

func (this *DriveFileServiceImpl) SearchDriveFilesShared(ctx crud.Context, query it.SearchDriveFilesSharedQuery) (result *it.SearchDriveFileResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "search drive files shared"); e != nil {
			err = e
		}
	}()
	result, err = crud.Search(ctx, crud.SearchParam[*domain.DriveFile, it.SearchDriveFilesSharedQuery, it.SearchDriveFileResult]{
		Action: "search drive files shared",
		Query:  query,
		SetQueryDefaults: func(q *it.SearchDriveFilesSharedQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: this.driveFileRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, q it.SearchDriveFilesSharedQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[*domain.DriveFile], error) {
			return this.driveFileRepo.GetDriveFilesSharedByUser(ctx, q.UserId, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *q.Page,
				Size:      *q.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchDriveFileResult {
			return &it.SearchDriveFileResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(paged *crud.PagedResult[*domain.DriveFile]) *it.SearchDriveFileResult {
			return &it.SearchDriveFileResult{Data: paged, HasData: paged.Items != nil}
		},
	})
	return result, err
}

func (this *DriveFileServiceImpl) GetDriveFileAncestors(ctx crud.Context, query it.GetDriveFileAncestorsQuery) (result *it.GetDriveFileAncestorsResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get drive file ancestors"); e != nil {
			err = e
		}
	}()
	vErrs := ft.NewValidationErrors()
	driveFile, err := this.assertDriveFileExists(ctx, query.ToDomainModel(), &vErrs)
	ft.PanicOnErr(err)

	if driveFile == nil {
		return &it.GetDriveFileAncestorsResult{ClientError: vErrs.ToClientError()}, nil
	}

	// Root has no parent: only owner roots are allowed.
	if driveFile.ParentDriveFileRef == nil {
		if driveFile.OwnerRef != nil && *driveFile.OwnerRef == query.UserId {
			return &it.GetDriveFileAncestorsResult{HasData: true, Data: []*domain.DriveFile{driveFile}}, nil
		}
		return &it.GetDriveFileAncestorsResult{
			ClientError: &fault.ClientError{
				Code:    "forbidden",
				Details: "drive file root is not allowed",
			},
		}, nil
	}

	ancestors, err := this.driveFileRepo.GetDriveFileParents(ctx, *driveFile.Id)
	ft.PanicOnErr(err)

	// Build map for O(1) lookup when walking up via parentRef.
	idToFile := make(map[model.Id]*domain.DriveFile, len(ancestors))
	for _, f := range ancestors {
		if f == nil || f.Id == nil {
			continue
		}
		idToFile[*f.Id] = f
	}

	// Walk upward from the original file until we find the first parent (ancestor)
	// that has at least `view` permission. We only return results up to that node.
	waiting := make([]*domain.DriveFile, 0)
	res := make([]*domain.DriveFile, 0)
	cur := driveFile
	for cur != nil && cur.ParentDriveFileRef != nil && *cur.ParentDriveFileRef != "" {
		parentId := *cur.ParentDriveFileRef
		parent := idToFile[parentId]
		if parent == nil {
			break
		}

		perm, err := this.resolvePermission(ctx, parent, query.UserId)
		ft.PanicOnErr(err)
		if perm < enum.DriveFileSharePermView {
			waiting = append(waiting, parent)
			cur = parent
			continue
		}

		// Found the first allowed ancestor. Append it and the waiting chain (from farthest to closest),
		// then append the original file.
		res = append(res, parent)
		for i := len(waiting) - 1; i >= 0; i-- {
			res = append(res, waiting[i])
		}
		res = append(res, driveFile)

		return &it.GetDriveFileAncestorsResult{HasData: true, Data: res}, nil
	}

	return &it.GetDriveFileAncestorsResult{
		ClientError: &fault.ClientError{
			Code:    "forbidden",
			Details: "drive file ancestors are not allowed",
		},
	}, nil
}
