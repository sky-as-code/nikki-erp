package services

import (
	"github.com/samber/lo"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file"
	driveFileAncestorIt "github.com/sky-as-code/nikki-erp/modules/drive2/interfaces/drive_file_ancestor"
)

type DriveFileDomainServiceImpl struct {
	logger logging.LoggerService
	config config.ConfigService

	driveFileRepo         it.DriveFileRepository
	driveFileAncestorRepo driveFileAncestorIt.DriveFileAncestorRepository
}

func NewDriveFileDomainService(
	logger logging.LoggerService,
	config config.ConfigService,

	driveFileRepo it.DriveFileRepository,
	driveFileAncestorRepo driveFileAncestorIt.DriveFileAncestorRepository,
) it.DriveFileDomainService {
	return &DriveFileDomainServiceImpl{
		logger:                logger,
		config:                config,
		driveFileRepo:         driveFileRepo,
		driveFileAncestorRepo: driveFileAncestorRepo,
	}
}

func (this *DriveFileDomainServiceImpl) CreateDriveFile(
	ctx corectx.Context, cmd it.CreateDriveFileCommand) (*it.CreateDriveFileResult, error) {
	return crud.Create(ctx, crud.CreateParam[models.DriveFile, *models.DriveFile]{
		Action:         "Create drive file",
		BaseRepoGetter: this.driveFileRepo,
		Data:           cmd.DriveFile,
	})
}

func (this *DriveFileDomainServiceImpl) UpdateDriveFileMetadata(ctx corectx.Context, cmd it.UpdateDriveFileMetadataCommand) (*it.UpdateDriveFileResult, error) {
	return crud.Update(ctx, crud.UpdateParam[models.DriveFile, *models.DriveFile]{
		Action:       "Update drive file metadata",
		DbRepoGetter: this.driveFileRepo,
		Data:         cmd,
		ValidateExtra: func(ctx corectx.Context, inputModel, foundModel *models.DriveFile, vErrs *ft.ClientErrors) error {
			if !lo.FromPtr(foundModel.GetIsFolder()) {
				namePtr := inputModel.GetName()
				if namePtr != nil {
					err := models.DriveFileNameValidate(*namePtr)
					if err != nil {
						vErrs.Append(*err)
					}
				}
			}
			return nil
		},
	})
}

func (this *DriveFileDomainServiceImpl) UpdateBulkDriveFileMetadata(ctx corectx.Context, cmd it.UpdateBulkDriveFileMetadataCommand) (*it.UpdateBulkDriveFileMetadataResult, error) {
	data := make([]models.DriveFile, 0, len(cmd))
	for _, item := range cmd {
		fieldData := item.GetFieldData()
		df := models.NewDriveFileFrom(fieldData)
		data = append(data, *df)
	}

	return crud.UpdateBulk(ctx, crud.UpdateBulkParam[models.DriveFile, *models.DriveFile, models.DriveFile]{
		Action:         "Update drive file metadata",
		BaseRepoGetter: this.driveFileRepo,
		Data:           data,
		ValidateExtra: func(ctx corectx.Context, inputModel, foundModel *models.DriveFile, vErrs *ft.ClientErrors) error {
			if !lo.FromPtr(foundModel.GetIsFolder()) {
				err := models.DriveFileNameValidate(lo.FromPtr(inputModel.GetName()))
				if err != nil {
					vErrs.Append(*err)
				}
			}
			return nil
		},
	})
}

func (this *DriveFileDomainServiceImpl) DeleteDriveFile(ctx corectx.Context, cmd it.DeleteDriveFileCommand) (*it.DeleteDriveFileResult, error) {
	return crud.DeleteOne(ctx, crud.DeleteOneParam{
		Action:                 "Delete Drive File",
		DbRepoGetter:           this.driveFileRepo,
		Cmd:                    dynamicmodel.DeleteOneCommand{},
		ValidateExtra:          nil,
		AfterValidationSuccess: nil,
	})
}

func (this *DriveFileDomainServiceImpl) DeleteDriveFiles(
	ctx corectx.Context, cmd it.DeleteDriveFilesCommand,
) (*it.DeleteDriveFilesResult, error) {
	if len(cmd.DriveFileIds) == 0 {
		return &it.DeleteDriveFilesResult{
			Data:    dynamicmodel.MutateResultData{AffectedCount: 0, AffectedAt: model.NewModelDateTime()},
			HasData: true,
		}, nil
	}

	deleted, err := this.driveFileRepo.DeleteByIds(ctx, cmd.DriveFileIds)
	if err != nil {
		return nil, err
	}

	return &it.DeleteDriveFilesResult{
		Data: dynamicmodel.MutateResultData{
			AffectedCount: deleted,
			AffectedAt:    model.NewModelDateTime(),
		},
		HasData: true,
	}, nil
}

func (this *DriveFileDomainServiceImpl) GetDriveFileAncestors(ctx corectx.Context, query it.GetDriveFileAncestorsQuery) (*it.GetDriveFileAncestorsResult, error) {
	panic("unimplemented")
}

func (this *DriveFileDomainServiceImpl) GetDriveFileById(ctx corectx.Context, query it.GetDriveFileByIdQuery) (*it.GetDriveFileByIdResult, error) {
	return crud.GetOne[models.DriveFile](ctx, crud.GetOneParam{
		Action:       "Get drive file by id",
		DbRepoGetter: this.driveFileRepo,
		Query: dynamicmodel.GetOneQuery{
			Id:     query.Id,
			Fields: query.Fields,
		},
	})
}

func (this *DriveFileDomainServiceImpl) GetDriveFileByParent(ctx corectx.Context, query it.GetDriveFileByParentQuery) (*it.GetDriveFileByParentResult, error) {
	panic("unimplemented")
}

func (this *DriveFileDomainServiceImpl) GetDriveFileChildren(
	ctx corectx.Context, query it.GetDriveFileChildrenQuery,
) (*it.GetDriveFileChildrenResult, error) {
	total, err := this.driveFileRepo.CountDriveFileChildren(ctx, query)
	if err != nil {
		return nil, err
	}

	items, err := this.driveFileRepo.GetDriveFileChildren(ctx, query)
	if err != nil {
		return nil, err
	}

	size := query.Size
	if size <= 0 {
		size = len(items)
	}

	return &it.GetDriveFileChildrenResult{
		Data: it.GetDriveFileChildrenResultData{
			Items: items,
			Total: total,
			Page:  query.Page,
			Size:  size,
		},
		HasData: len(items) > 0,
	}, nil
}

func (this *DriveFileDomainServiceImpl) MoveDriveFile(ctx corectx.Context, cmd it.MoveDriveFileCommand) (*it.MoveDriveFileResult, error) {
	panic("unimplemented")
}

func (this *DriveFileDomainServiceImpl) MoveDriveFileToTrash(ctx corectx.Context, cmd it.MoveDriveFileToTrashCommand) (*it.MoveDriveFileToTrashResult, error) {
	status := models.DriveFileStatusInTrash
	driveFile := models.NewDriveFile()
	driveFile.SetId(&cmd.DriveFileId)
	driveFile.SetEtag(&cmd.Etag)
	driveFile.SetStatus(&status)
	return this.UpdateDriveFileMetadata(ctx, *driveFile)
}

func (this *DriveFileDomainServiceImpl) RestoreDriveFile(ctx corectx.Context, cmd it.RestoreDriveFileCommand) (*it.RestoreDriveFileResult, error) {
	panic("unimplemented")
}

func (this *DriveFileDomainServiceImpl) SearchDriveFile(ctx corectx.Context, query it.SearchDriveFileQuery) (*it.SearchDriveFileResult, error) {
	panic("unimplemented")
}
