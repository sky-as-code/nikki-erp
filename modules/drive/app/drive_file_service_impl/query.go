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

			permission, err := this.permissionSvc.ResolvePermission(ctx, driveFile, q.UserId)
			ft.PanicOnErr(err)
			if !permission.CanView() {
				vErrs.AppendNotAllowed("driveFileId", "drive file")
				return nil, nil
			}
			return driveFile, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetDriveFileByIdResult {
			return &it.GetDriveFileByIdResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(d *domain.DriveFile) *it.GetDriveFileByIdResult {
			ft.PanicOnErr(this.enrichDriveFilesWithOwners(ctx, []*domain.DriveFile{d}))
			ft.PanicOnErr(this.permissionSvc.EnrichDriveFilesWithPermissions(ctx, []*domain.DriveFile{d}, query.UserId))
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

	if driveFile.IsFolder {
		return nil, nil, &fault.ClientError{
			Code:    "bad_request",
			Details: fault.ValidationErrors{"driveFileId": "cannot download a folder"},
		}
	}

	// Permission check skipped here temporarily (public stream route); restore resolvePermission + CanView when auth/token is in place.

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

		permission, err := this.permissionSvc.ResolvePermission(ctx, parent, query.UserId)
		ft.PanicOnErr(err)
		if !permission.CanView() {
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
			if paged != nil {
				ft.PanicOnErr(this.enrichDriveFilesWithOwners(ctx, paged.Items))
				ft.PanicOnErr(this.permissionSvc.EnrichDriveFilesWithPermissions(ctx, paged.Items, query.UserId))
			}
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
			if paged != nil {
				ft.PanicOnErr(this.enrichDriveFilesWithOwners(ctx, paged.Items))
				ft.PanicOnErr(this.permissionSvc.EnrichDriveFilesWithPermissions(ctx, paged.Items, query.UserId))
			}
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
			if paged != nil {
				ft.PanicOnErr(this.enrichDriveFilesWithOwners(ctx, paged.Items))
				ft.PanicOnErr(this.permissionSvc.EnrichDriveFilesWithPermissions(ctx, paged.Items, query.UserId))
			}
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
			if paged != nil {
				ft.PanicOnErr(this.enrichDriveFilesWithOwners(ctx, paged.Items))
				ft.PanicOnErr(this.permissionSvc.EnrichDriveFilesWithPermissions(ctx, paged.Items, query.UserId))
			}
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
		// Root may be viewable either by owner OR by direct share.
		// Use the same permission resolver for consistent behavior.
		perm, err := this.permissionSvc.ResolvePermission(ctx, driveFile, query.UserId)
		ft.PanicOnErr(err)
		if !perm.CanView() {
			return &it.GetDriveFileAncestorsResult{HasData: true, Data: []*domain.DriveFile{}}, nil
		}

		data := []*domain.DriveFile{driveFile}
		ft.PanicOnErr(this.enrichDriveFilesWithOwners(ctx, data))
		ft.PanicOnErr(this.permissionSvc.EnrichDriveFilesWithPermissions(ctx, data, query.UserId))
		return &it.GetDriveFileAncestorsResult{HasData: true, Data: data}, nil
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

	// Walk upward and keep track of the farthest node (highest ancestor)
	// that has at least `view` permission. Only return 403 when no node
	// in the chain can be viewed.
	path := make([]*domain.DriveFile, 0, len(ancestors)+1)
	path = append(path, driveFile) // index 0 = the original file

	highestViewIdx := -1
	selfPerm, err := this.permissionSvc.ResolvePermission(ctx, driveFile, query.UserId)
	ft.PanicOnErr(err)
	if selfPerm.CanView() {
		highestViewIdx = 0
	}

	cur := driveFile
	for cur != nil && cur.ParentDriveFileRef != nil && *cur.ParentDriveFileRef != "" {
		parentId := *cur.ParentDriveFileRef
		parent := idToFile[parentId]
		if parent == nil {
			break
		}

		path = append(path, parent)
		perm, err := this.permissionSvc.ResolvePermission(ctx, parent, query.UserId)
		ft.PanicOnErr(err)
		if perm.CanView() {
			highestViewIdx = len(path) - 1
		}

		cur = parent
	}

	if highestViewIdx == -1 {
		return &it.GetDriveFileAncestorsResult{
			ClientError: &fault.ClientError{
				Code:    "forbidden",
				Details: "drive file ancestors are not allowed",
			},
		}, nil
	}

	// res order: from highest view ancestor down to the original file.
	res := make([]*domain.DriveFile, 0, highestViewIdx+1)
	for i := highestViewIdx; i >= 0; i-- {
		res = append(res, path[i])
	}

	ft.PanicOnErr(this.enrichDriveFilesWithOwners(ctx, res))
	ft.PanicOnErr(this.permissionSvc.EnrichDriveFilesWithPermissions(ctx, res, query.UserId))
	return &it.GetDriveFileAncestorsResult{HasData: true, Data: res}, nil
}
