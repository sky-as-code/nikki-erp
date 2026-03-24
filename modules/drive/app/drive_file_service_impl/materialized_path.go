package drive_file_service_impl

import (
	"errors"
	"strings"

	"github.com/samber/lo"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
)

var errParentDriveFileNotFound = errors.New("parent drive file not found")

func (this *DriveFileServiceImpl) buildMaterializedPathUnderParent(ctx crud.Context, parentRef *model.Id, selfId model.Id) (string, error) {
	if parentRef == nil || *parentRef == "" {
		return materializedPathFromIds([]model.Id{selfId}), nil
	}
	ancestors, err := this.driveFileRepo.GetDriveFileParents(ctx, *parentRef)
	if err != nil {
		return "", err
	}
	if len(ancestors) == 0 {
		return "", errParentDriveFileNotFound
	}
	chain := make([]*domain.DriveFile, len(ancestors))
	copy(chain, ancestors)
	lo.Reverse(chain)
	ids := make([]model.Id, 0, len(chain)+1)
	for _, f := range chain {
		if f != nil && f.Id != nil {
			ids = append(ids, *f.Id)
		}
	}
	ids = append(ids, selfId)
	return materializedPathFromIds(ids), nil
}

func (this *DriveFileServiceImpl) buildMaterializedPathForFile(ctx crud.Context, file *domain.DriveFile) (string, error) {
	if file == nil || file.Id == nil {
		return "", nil
	}
	ancestors, err := this.driveFileRepo.GetDriveFileParents(ctx, *file.Id)
	if err != nil {
		return "", err
	}
	chain := make([]*domain.DriveFile, len(ancestors))
	copy(chain, ancestors)
	lo.Reverse(chain)
	ids := make([]model.Id, 0, len(chain))
	for _, f := range chain {
		if f != nil && f.Id != nil {
			ids = append(ids, *f.Id)
		}
	}
	return materializedPathFromIds(ids), nil
}

// propagateMaterializedPathsAfterMove loads the full subtree (GetDriveFileChildren + BuildTree), then pre-order Overwrite for each descendant. Root path is already persisted.
func (this *DriveFileServiceImpl) propagateMaterializedPathsAfterMove(ctx crud.Context, root *domain.DriveFile, rootPath string) error {
	if root == nil || root.Id == nil || !root.IsFolder || rootPath == "" {
		return nil
	}
	descendants, err := this.driveFileRepo.GetDriveFileChildren(ctx, *root.Id)
	if err != nil {
		return err
	}
	root.BuildTree(descendants)
	return this.preOrderOverwriteMaterializedPaths(ctx, root, rootPath, true)
}

func (this *DriveFileServiceImpl) preOrderOverwriteMaterializedPaths(ctx crud.Context, node *domain.DriveFile, pathForNode string, isRoot bool) error {
	if node == nil || node.Id == nil || node.Etag == nil {
		return nil
	}
	if !isRoot {
		p := pathForNode
		node.MaterializedPath = &p
		updated, err := this.driveFileRepo.Overwrite(ctx, node, *node.Etag)
		if err != nil {
			return err
		}
		node.Etag = updated.Etag
	}
	for _, child := range node.Children {
		if child == nil || child.Id == nil {
			continue
		}
		childPath := joinMaterializedPathChild(pathForNode, *child.Id)
		if err := this.preOrderOverwriteMaterializedPaths(ctx, child, childPath, false); err != nil {
			return err
		}
	}
	return nil
}

func materializedPathFromIds(ids []model.Id) string {
	if len(ids) == 0 {
		return ""
	}
	var b strings.Builder
	for _, id := range ids {
		b.WriteByte('/')
		b.WriteString(string(id))
	}
	return b.String()
}

func joinMaterializedPathChild(parentPath string, childId model.Id) string {
	parentPath = strings.TrimSuffix(parentPath, "/")
	return parentPath + "/" + string(childId)
}
