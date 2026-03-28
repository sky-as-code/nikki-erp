package drive_file_service_impl

import (
	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

func (this *DriveFileServiceImpl) recalculateSizeOfParent(ctx crud.Context,
	parentId model.Id, sizeDelta uint64, inc bool) error {
	driveFiles, err := this.driveFileRepo.GetDriveFileParents(ctx, parentId)
	ft.PanicOnErr(err)

	updateCmds := make([]it.UpdateDriveFileMetadataCommand, 0, len(driveFiles))
	for _, f := range driveFiles {
		cmd := it.UpdateDriveFileMetadataCommand{
			Id:   *f.Id,
			Etag: *f.Etag,
		}

		if inc {
			cmd.Size = f.Size + uint64(sizeDelta)
		} else {
			cmd.Size = f.Size - uint64(sizeDelta)
		}

		updateCmds = append(updateCmds, cmd)
	}

	_, err = this.UpdateBulkDriveFileMetadata(ctx, it.UpdateBulkDriveFileMetadataCommand{
		DriveFiles: updateCmds,
	})
	ft.PanicOnErr(err)

	return nil
}

func (this *DriveFileServiceImpl) updateChildrenStatus(ctx crud.Context, driveFile *domain.DriveFile, status enum.DriveFileStatus) error {
	children, err := this.driveFileRepo.GetDriveFileChildren(ctx, *driveFile.Id)
	ft.PanicOnErr(err)

	updateCmds := make([]it.UpdateDriveFileMetadataCommand, 0, len(children))
	for _, f := range children {
		updateCmds = append(updateCmds, it.UpdateDriveFileMetadataCommand{
			Id:     *f.Id,
			Etag:   *f.Etag,
			Status: status,
		})
	}

	_, err = this.UpdateBulkDriveFileMetadata(ctx, it.UpdateBulkDriveFileMetadataCommand{
		DriveFiles: updateCmds,
	})
	ft.PanicOnErr(err)

	return nil
}

// insertAncestorsForFile builds the ancestor list from the parent's existing ancestors
// plus the parent itself, then inserts rows for the new file.
func (this *DriveFileServiceImpl) insertAncestorsForFile(ctx crud.Context, fileId model.Id, parentRef *model.Id) error {
	if parentRef == nil || *parentRef == "" {
		return nil
	}
	parentAncestors, err := this.driveFileRepo.GetAncestorIds(ctx, *parentRef)
	if err != nil {
		return err
	}
	// ancestors = parent's ancestors (root first) + parent itself
	ancestorIds := append(parentAncestors, *parentRef)
	return this.driveFileRepo.InsertAncestors(ctx, fileId, ancestorIds)
}

// rebuildAncestorsForSubtree deletes old ancestor rows for the entire subtree
// and re-inserts them based on the new parent hierarchy.
func (this *DriveFileServiceImpl) rebuildAncestorsForSubtree(ctx crud.Context, root *domain.DriveFile, newParentRef *model.Id) error {
	if root == nil || root.Id == nil {
		return nil
	}
	children, err := this.driveFileRepo.GetDriveFileChildren(ctx, *root.Id)
	if err != nil {
		return err
	}

	allFileIds := make([]model.Id, 0, len(children)+1)
	allFileIds = append(allFileIds, *root.Id)
	for _, c := range children {
		if c != nil && c.Id != nil {
			allFileIds = append(allFileIds, *c.Id)
		}
	}

	if err := this.driveFileRepo.DeleteAncestorsByFileIds(ctx, allFileIds); err != nil {
		return err
	}

	var newRootAncestors []model.Id
	if newParentRef != nil && *newParentRef != "" {
		parentAncestors, err := this.driveFileRepo.GetAncestorIds(ctx, *newParentRef)
		if err != nil {
			return err
		}
		newRootAncestors = append(parentAncestors, *newParentRef)
	}

	if err := this.driveFileRepo.InsertAncestors(ctx, *root.Id, newRootAncestors); err != nil {
		return err
	}

	if !root.IsFolder || len(children) == 0 {
		return nil
	}

	root.BuildTree(children)
	// rootAncestors = newRootAncestors + root itself
	rootAncestors := make([]model.Id, len(newRootAncestors)+1)
	copy(rootAncestors, newRootAncestors)
	rootAncestors[len(newRootAncestors)] = *root.Id
	return this.insertAncestorsPreOrder(ctx, root, rootAncestors)
}

func (this *DriveFileServiceImpl) insertAncestorsPreOrder(ctx crud.Context, node *domain.DriveFile, ancestorsOfChildren []model.Id) error {
	for _, child := range node.Children {
		if child == nil || child.Id == nil {
			continue
		}
		if err := this.driveFileRepo.InsertAncestors(ctx, *child.Id, ancestorsOfChildren); err != nil {
			return err
		}
		if child.IsFolder && len(child.Children) > 0 {
			childAncestors := make([]model.Id, len(ancestorsOfChildren)+1)
			copy(childAncestors, ancestorsOfChildren)
			childAncestors[len(ancestorsOfChildren)] = *child.Id
			if err := this.insertAncestorsPreOrder(ctx, child, childAncestors); err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *DriveFileServiceImpl) sanitizeDriveFile(d *domain.DriveFile) {
	if d != nil && d.Name != "" {
		d.Name = defense.SanitizePlainText(d.Name, true)
	}
}

func (this *DriveFileServiceImpl) assertDriveFileExists(ctx crud.Context, d *domain.DriveFile, vErrs *ft.ValidationErrors) (*domain.DriveFile, error) {
	if d == nil || d.Id == nil {
		return nil, nil
	}

	driveFile, err := this.driveFileRepo.FindById(ctx, *d.Id)
	ft.PanicOnErr(err)

	if driveFile == nil {
		if vErrs != nil {
			vErrs.Append("driveFileId", "drive file not found")
		}
		return nil, nil
	}

	return driveFile, nil
}
