package drive_file_ancestor

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain/models"
)

func init() {
	var req cqrs.Request
	req = (*CreateDriveFileAncestorCommand)(nil)
	req = (*CreateBulkDriveFileAncestorsCommand)(nil)
	req = (*DeleteDriveFileAncestorCommand)(nil)
	req = (*GetDriveFileAncestorQuery)(nil)
	req = (*DriveFileAncestorExistsQuery)(nil)
	req = (*SearchDriveFileAncestorsQuery)(nil)
	req = (*UpdateDriveFileAncestorCommand)(nil)
	util.Unused(req)
}

var createDriveFileAncestorCommandType = cqrs.RequestType{
	Module:    "drive2",
	Submodule: "drive_file_ancestor",
	Action:    "createDriveFileAncestor",
}

type CreateDriveFileAncestorCommand struct {
	models.DriveFileAncestor
}

func (CreateDriveFileAncestorCommand) CqrsRequestType() cqrs.RequestType {
	return createDriveFileAncestorCommandType
}

func (CreateDriveFileAncestorCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.DriveFileAncestorSchemaName)
}

type CreateDriveFileAncestorResult = dyn.OpResult[models.DriveFileAncestor]

var createBulkDriveFileAncestorsCommandType = cqrs.RequestType{
	Module:    "drive2",
	Submodule: "drive_file_ancestor",
	Action:    "createBulkDriveFileAncestors",
}

type CreateBulkDriveFileAncestorsCommand struct {
	Items []models.DriveFileAncestor `json:"items"`
}

func (CreateBulkDriveFileAncestorsCommand) CqrsRequestType() cqrs.RequestType {
	return createBulkDriveFileAncestorsCommandType
}

type CreateBulkDriveFileAncestorsResult = dyn.OpResult[[]models.DriveFileAncestor]

var deleteDriveFileAncestorCommandType = cqrs.RequestType{
	Module:    "drive2",
	Submodule: "drive_file_ancestor",
	Action:    "deleteDriveFileAncestor",
}

type DeleteDriveFileAncestorCommand dyn.DeleteOneCommand

func (DeleteDriveFileAncestorCommand) CqrsRequestType() cqrs.RequestType {
	return deleteDriveFileAncestorCommandType
}

type DeleteDriveFileAncestorResult = dyn.OpResult[dyn.MutateResultData]

var getDriveFileAncestorQueryType = cqrs.RequestType{
	Module:    "drive2",
	Submodule: "drive_file_ancestor",
	Action:    "getDriveFileAncestor",
}

type GetDriveFileAncestorQuery dyn.GetOneQuery

func (GetDriveFileAncestorQuery) CqrsRequestType() cqrs.RequestType {
	return getDriveFileAncestorQueryType
}

type GetDriveFileAncestorResult = dyn.OpResult[models.DriveFileAncestor]

var driveFileAncestorExistsQueryType = cqrs.RequestType{
	Module:    "drive2",
	Submodule: "drive_file_ancestor",
	Action:    "driveFileAncestorExists",
}

type DriveFileAncestorExistsQuery dyn.ExistsQuery

func (DriveFileAncestorExistsQuery) CqrsRequestType() cqrs.RequestType {
	return driveFileAncestorExistsQueryType
}

type DriveFileAncestorExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchDriveFileAncestorsQueryType = cqrs.RequestType{
	Module:    "drive2",
	Submodule: "drive_file_ancestor",
	Action:    "searchDriveFileAncestors",
}

type SearchDriveFileAncestorsQuery dyn.SearchQuery

func (SearchDriveFileAncestorsQuery) CqrsRequestType() cqrs.RequestType {
	return searchDriveFileAncestorsQueryType
}

type SearchDriveFileAncestorsResultData = dyn.PagedResultData[models.DriveFileAncestor]
type SearchDriveFileAncestorsResult = dyn.OpResult[SearchDriveFileAncestorsResultData]

var updateDriveFileAncestorCommandType = cqrs.RequestType{
	Module:    "drive2",
	Submodule: "drive_file_ancestor",
	Action:    "updateDriveFileAncestor",
}

type UpdateDriveFileAncestorCommand struct {
	models.DriveFileAncestor
}

func (UpdateDriveFileAncestorCommand) CqrsRequestType() cqrs.RequestType {
	return updateDriveFileAncestorCommandType
}

func (UpdateDriveFileAncestorCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(models.DriveFileAncestorSchemaName)
}

type UpdateDriveFileAncestorResult = dyn.OpResult[dyn.MutateResultData]
