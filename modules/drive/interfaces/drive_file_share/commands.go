package drive_file_share

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
)

type CreateDriveFileShareCommand struct {
	FileRef    model.Id                `param:"driveFileId"`
	UserRef    model.Id                `json:"userRef"`
	Permission enum.DriveFileSharePerm `json:"permission"`
}

type CreateDriveFileShareResult = crud.OpResult[*domain.DriveFileShare]

type CreateBulkDriveFileShareCommand struct {
	FileRef    model.Id                `param:"driveFileId"`
	UserRefs   []model.Id              `json:"userRefs"`
	Permission enum.DriveFileSharePerm `json:"permission"`
}

type CreateBulkDriveFileShareResult = crud.OpResult[[]*domain.DriveFileShare]

type UpdateDriveFileShareCommand struct {
	Id         model.Id                `param:"driveFileShareId" json:"driveFileShareId"`
	Permission enum.DriveFileSharePerm `json:"permission"`
}

type UpdateDriveFileShareResult = crud.OpResult[*domain.DriveFileShare]

type GetDriveFileShareByIdQuery struct {
	DriveFileShareId model.Id `param:"driveFileShareId" json:"driveFileShareId"`
}

type GetDriveFileShareByIdResult = crud.OpResult[*domain.DriveFileShare]

type GetDriveFileShareByFileIdQuery struct {
	DriveFileId model.Id `param:"driveFileId" json:"driveFileId"`
}

type GetDriveFileShareByFileIdResultData = crud.PagedResult[*domain.DriveFileShare]
type GetDriveFileShareByFileIdResult = crud.OpResult[*GetDriveFileShareByFileIdResultData]

type GetDriveFileShareByUserQuery struct {
	UserId model.Id `param:"userId" json:"user_id"`
}

type GetDriveFileShareByUserResultData = crud.PagedResult[*domain.DriveFileShare]
type GetDriveFileShareByUserResult = crud.OpResult[*GetDriveFileShareByUserResultData]

type SearchDriveFileShareQuery struct {
	crud.SearchQuery
}

type SearchDriveFileShareResultData = crud.PagedResult[*domain.DriveFileShare]
type SearchDriveFileShareResult = crud.OpResult[*SearchDriveFileShareResultData]

type DeleteDriveFileShareCommand struct {
	DriveFileShareId model.Id `param:"id" json:"id"`
}

type DeleteDriveFileShareResult = crud.DeletionResult
