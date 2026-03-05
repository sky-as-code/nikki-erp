package repository

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	"github.com/sky-as-code/nikki-erp/modules/drive/infra/ent"
)

// --- DriveFileShare ---

func entToDriveFileShare(e *ent.DriveFileShare) *domain.DriveFileShare {
	if e == nil {
		return nil
	}
	d := &domain.DriveFileShare{}
	d.Id = &e.ID
	d.Etag = &e.Etag
	d.CreatedAt = &e.CreatedAt
	d.UpdatedAt = &e.UpdatedAt
	d.FileRef = model.Id(e.FileRef)
	d.UserRef = model.Id(e.UserRef)
	d.Permission = enum.DriveFileSharePermValue[e.Permission]
	if d.Permission == 0 {
		d.Permission = enum.DriveFileSharePermDefault
	}
	return d
}

func entToDriveFileShares(entities []*ent.DriveFileShare) []*domain.DriveFileShare {
	if entities == nil {
		return nil
	}
	result := make([]*domain.DriveFileShare, len(entities))
	for i, e := range entities {
		result[i] = entToDriveFileShare(e)
	}
	return result
}

// --- DriveFile ---

func entToDriveFile(e *ent.DriveFile) *domain.DriveFile {
	if e == nil {
		return nil
	}
	d := &domain.DriveFile{}
	d.Id = &e.ID
	d.Etag = &e.Etag
	d.CreatedAt = &e.CreatedAt
	d.UpdatedAt = &e.UpdatedAt
	if !e.DeletedAt.Equal(deletedAtNotDeleted) {
		d.DeletedAt = &e.DeletedAt
	}
	ownerRef := model.Id(e.OwnerRef)
	d.OwnerRef = &ownerRef
	if e.ParentFileRef != nil {
		parentRef := model.Id(*e.ParentFileRef)
		d.ParentDriveFileRef = &parentRef
	}
	d.Name = e.Name
	d.MINE = e.Mime
	d.IsFolder = e.IsFolder
	d.Size = uint64(e.Size)
	d.Path = e.Path
	d.Storage = enum.DriveFileStorageValue[e.Storage]
	if d.Storage == 0 {
		d.Storage = enum.DriveFileStorageDefault
	}
	d.Visibility = enum.DriveFileVisibilityValue[e.Visibility]
	if d.Visibility == 0 {
		d.Visibility = enum.DriveFileVisibilityDefault
	}
	return d
}

func entToDriveFiles(entities []*ent.DriveFile) []*domain.DriveFile {
	if entities == nil {
		return nil
	}
	result := make([]*domain.DriveFile, len(entities))
	for i, e := range entities {
		result[i] = entToDriveFile(e)
	}
	return result
}
