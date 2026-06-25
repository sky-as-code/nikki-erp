package drive_file_ancestor

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
)

type DriveFileAncestorDomainService interface {
	CreateDriveFileAncestor(
		ctx corectx.Context, cmd CreateDriveFileAncestorCommand,
		opts ...corecrud.ServiceCreateOptions[*models.DriveFileAncestor],
	) (*CreateDriveFileAncestorResult, error)
	CreateBulkDriveFileAncestors(
		ctx corectx.Context, cmd CreateBulkDriveFileAncestorsCommand,
	) (*CreateBulkDriveFileAncestorsResult, error)
	DeleteDriveFileAncestor(
		ctx corectx.Context, cmd DeleteDriveFileAncestorCommand, opts ...corecrud.ServiceDeleteOptions,
	) (*DeleteDriveFileAncestorResult, error)
	DriveFileAncestorExists(ctx corectx.Context, query DriveFileAncestorExistsQuery) (*DriveFileAncestorExistsResult, error)
	GetDriveFileAncestor(ctx corectx.Context, query GetDriveFileAncestorQuery) (*dyn.OpResult[models.DriveFileAncestor], error)
	SearchDriveFileAncestors(
		ctx corectx.Context, query SearchDriveFileAncestorsQuery, opts ...corecrud.ServiceSearchOptions,
	) (*SearchDriveFileAncestorsResult, error)
	UpdateDriveFileAncestor(
		ctx corectx.Context, cmd UpdateDriveFileAncestorCommand,
		opts ...corecrud.ServiceUpdateOptions[*models.DriveFileAncestor],
	) (*UpdateDriveFileAncestorResult, error)
}
