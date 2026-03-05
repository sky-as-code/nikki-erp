package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter/cqrs_bus/identity_cqrs"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	driFileIt "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

type DriveFileShareServiceImpl struct {
	driveFileShareRepo it.DriveFileShareRepository
	driveFileRepo      driFileIt.DriveFileRepository
	identityCqrs       identity_cqrs.IdentityCqrsAdapter
}

func NewDriveFileShareService(driveFileShareRepo it.DriveFileShareRepository, driveFileRepo driFileIt.DriveFileRepository, identityCqrsAdapter identity_cqrs.IdentityCqrsAdapter) it.DriveFileShareService {
	return &DriveFileShareServiceImpl{
		driveFileShareRepo: driveFileShareRepo,
		driveFileRepo:      driveFileRepo,
		identityCqrs:       identityCqrsAdapter,
	}
}

// Create

func (this *DriveFileShareServiceImpl) CreateDriveFileShare(ctx crud.Context, cmd it.CreateDriveFileShareCommand) (
	*it.CreateDriveFileShareResult, error) {
	return crud.Create(ctx, crud.CreateParam[*domain.DriveFileShare, *it.CreateDriveFileShareCommand, it.CreateDriveFileShareResult]{
		Action:              "create drive file share",
		Command:             &cmd,
		AssertBusinessRules: this.assertCreateDriveFileShareBussinessRules,
		RepoCreate:          this.driveFileShareRepo.Create,
		SetDefault:          this.setCreateDefaults,
		Sanitize:            func(d *domain.DriveFileShare) {},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateDriveFileShareResult {
			return &it.CreateDriveFileShareResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(d *domain.DriveFileShare) *it.CreateDriveFileShareResult {
			return &it.CreateDriveFileShareResult{HasData: true, Data: d}
		},
	})
}

// Create Bulk

func (this *DriveFileShareServiceImpl) CreateBulkDriveFileShares(ctx crud.Context, cmd it.CreateBulkDriveFileShareCommand) (
	*it.CreateBulkDriveFileShareResult, error) {
	return crud.CreateBulk(ctx, crud.CreateBulkParam[*domain.DriveFileShare, it.CreateBulkDriveFileShareCommand, it.CreateBulkDriveFileShareResult]{
		Action:         "create bulk drive file shares",
		Command:        cmd,
		RepoCreateBulk: this.repoCreateBulk,
		SetDefault:     this.setCreateDefaults,
		Sanitize:       func(d *domain.DriveFileShare) {},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateBulkDriveFileShareResult {
			return &it.CreateBulkDriveFileShareResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(models []*domain.DriveFileShare) *it.CreateBulkDriveFileShareResult {
			return &it.CreateBulkDriveFileShareResult{HasData: models != nil, Data: models}
		},
	})
}

func (this *DriveFileShareServiceImpl) repoCreateBulk(ctx crud.Context, models []*domain.DriveFileShare) ([]*domain.DriveFileShare, error) {
	created := make([]*domain.DriveFileShare, 0, len(models))
	for _, m := range models {
		result, err := this.driveFileShareRepo.Create(ctx, m)
		if err != nil {
			return nil, err
		}
		created = append(created, result)
	}
	return created, nil
}

// Update

func (this *DriveFileShareServiceImpl) UpdateDriveFileShare(ctx crud.Context, cmd it.UpdateDriveFileShareCommand) (
	*it.UpdateDriveFileShareResult, error) {
	return crud.Update(ctx, crud.UpdateParam[*domain.DriveFileShare, it.UpdateDriveFileShareCommand, it.UpdateDriveFileShareResult]{
		Action:       "update drive file share",
		Command:      cmd,
		AssertExists: this.assertDriveFileShareExists,
		RepoUpdate:   this.driveFileShareRepo.Update,
		Sanitize:     func(d *domain.DriveFileShare) {},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateDriveFileShareResult {
			return &it.UpdateDriveFileShareResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(d *domain.DriveFileShare) *it.UpdateDriveFileShareResult {
			return &it.UpdateDriveFileShareResult{HasData: true, Data: d}
		},
	})
}

// Get by ID

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
			return this.driveFileShareRepo.ListByFileRef(ctx, it.ListByFileRefParam{
				FileRef: q.DriveFileId,
				SearchParam: it.SearchParam{
					Predicate: predicate,
					Order:     order,
					Page:      *q.Page,
					Size:      *q.Size,
				},
			})
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
			return this.driveFileShareRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *q.Page,
				Size:      *q.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchDriveFileShareResult {
			return &it.SearchDriveFileShareResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(paged *crud.PagedResult[*domain.DriveFileShare]) *it.SearchDriveFileShareResult {
			return &it.SearchDriveFileShareResult{Data: paged, HasData: paged.Items != nil}
		},
	})
}

// Delete

func (this *DriveFileShareServiceImpl) DeleteDriveFileShare(ctx crud.Context, cmd it.DeleteDriveFileShareCommand) (
	*it.DeleteDriveFileShareResult, error) {
	return crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.DriveFileShare, it.DeleteDriveFileShareCommand, it.DeleteDriveFileShareResult]{
		Action:       "delete drive file share",
		Command:      cmd,
		AssertExists: this.assertDriveFileShareExists,
		RepoDelete: func(ctx crud.Context, d *domain.DriveFileShare) (int, error) {
			return this.driveFileShareRepo.DeleteById(ctx, *d.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.DeleteDriveFileShareResult {
			return &it.DeleteDriveFileShareResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(_ *domain.DriveFileShare, deletedCount int,
		) *it.DeleteDriveFileShareResult {
			return crud.NewSuccessDeletionResult(cmd.DriveFileShareId, &deletedCount)
		},
	})
}

// --- helpers ---

func (this *DriveFileShareServiceImpl) assertCreateDriveFileShareBussinessRules(ctx crud.Context,
	d *domain.DriveFileShare, vErrs *ft.ValidationErrors) error {

	// assert file ref exists
	driveFile, err := this.driveFileRepo.FindById(ctx, d.FileRef)
	if err != nil {
		return err
	}

	if driveFile == nil {
		if vErrs != nil {
			vErrs.Append("driveFileId", "drive file not found")
		}
		return nil
	}

	// assert user ref exists
	exists, err, clientErr := this.identityCqrs.UserExists(ctx, d.UserRef)
	if err != nil {
		return err
	}

	if clientErr != nil {
		if vErrs != nil {
			vErrs.MergeClientError(clientErr)
		}

		return nil
	}

	if !exists {
		if vErrs != nil {
			vErrs.Append("userRef", "user not found")
		}

		return nil
	}

	return nil
}

func (this *DriveFileShareServiceImpl) setCreateDefaults(d *domain.DriveFileShare) {
	d.SetDefaults()
	if d.Permission == 0 {
		d.Permission = enum.DriveFileSharePermDefault
	}
}

func (this *DriveFileShareServiceImpl) assertDriveFileShareExists(ctx crud.Context, d *domain.DriveFileShare, vErrs *ft.ValidationErrors) (*domain.DriveFileShare, error) {
	if d == nil || d.Id == nil {
		return nil, nil
	}
	share, err := this.driveFileShareRepo.FindById(ctx, *d.Id)
	if err != nil {
		return nil, err
	}
	if share == nil {
		if vErrs != nil {
			vErrs.AppendNotFound("driveFileShareId", "drive file share")
		}
		return nil, nil
	}
	return share, nil
}

func (this *DriveFileShareServiceImpl) sanitizeDriveFileShare(d *domain.DriveFileShare) {
	if d == nil {
		return
	}
	if d.Permission == 0 {
		d.Permission = enum.DriveFileSharePermDefault
	}
}
