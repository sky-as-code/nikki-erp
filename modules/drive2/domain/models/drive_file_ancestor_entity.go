package models

import (
	"math"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	DriveFileAncestorSchemaName = "drive2.drive_file_ancestor"

	DriveFileAncestorFieldId          = basemodel.FieldId
	DriveFileAncestorFieldFileRef     = "file_ref"
	DriveFileAncestorFieldAncestorRef = "ancestor_ref"
	DriveFileAncestorFieldDepth       = "depth"

	DriveFileAncestorEdgeDriveFile    = "drive_file"
	DriveFileAncestorEdgeAncestorFile = "ancestor_file"
)

func DriveFileAncestorSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(DriveFileAncestorSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Drive file ancestor"}).
		TableName("dri_file_ancestors").
		CompositeUnique(DriveFileAncestorFieldFileRef, DriveFileAncestorFieldAncestorRef).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(DriveFileAncestorFieldFileRef).
				Label(model.LangJson{model.LanguageCodeEnUs: "File"}).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(DriveFileAncestorFieldAncestorRef).
				Label(model.LangJson{model.LanguageCodeEnUs: "Ancestor file"}).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(DriveFileAncestorFieldDepth).
				Label(model.LangJson{model.LanguageCodeEnUs: "Depth"}).
				DataType(dmodel.FieldDataTypeInt64(0, math.MaxInt32)).
				RequiredForCreate().
				Default(int64(0)),
		).
		SearchIndexGroup(dmodel.SearchIndexGroupParam{
			IndexName: "drivefileancestor_ancestor_ref",
			Fields:    []string{DriveFileAncestorFieldAncestorRef},
		}).
		EdgeTo(
			dmodel.Edge(DriveFileAncestorEdgeDriveFile).
				Label(model.LangJson{model.LanguageCodeEnUs: "Drive file"}).
				ManyToOne(DriveFileSchemaName, dmodel.DynamicFields{
					DriveFileAncestorFieldFileRef: DriveFileFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(DriveFileAncestorEdgeAncestorFile).
				Label(model.LangJson{model.LanguageCodeEnUs: "Ancestor file"}).
				ManyToOne(DriveFileSchemaName, dmodel.DynamicFields{
					DriveFileAncestorFieldAncestorRef: DriveFileFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

type DriveFileAncestor struct {
	basemodel.DynamicModelBase
}

func NewDriveFileAncestor() *DriveFileAncestor {
	return &DriveFileAncestor{basemodel.NewDynamicModel()}
}

func NewDriveFileAncestorFrom(src dmodel.DynamicFields) *DriveFileAncestor {
	return &DriveFileAncestor{basemodel.NewDynamicModel(src)}
}

func (this DriveFileAncestor) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *DriveFileAncestor) SetId(v *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, v)
}

func (this DriveFileAncestor) GetFileRef() *model.Id {
	return this.GetFieldData().GetModelId(DriveFileAncestorFieldFileRef)
}

func (this *DriveFileAncestor) SetFileRef(v *model.Id) {
	this.GetFieldData().SetModelId(DriveFileAncestorFieldFileRef, v)
}

func (this DriveFileAncestor) GetAncestorRef() *model.Id {
	return this.GetFieldData().GetModelId(DriveFileAncestorFieldAncestorRef)
}

func (this *DriveFileAncestor) SetAncestorRef(v *model.Id) {
	this.GetFieldData().SetModelId(DriveFileAncestorFieldAncestorRef, v)
}

func (this DriveFileAncestor) GetDepth() *int64 {
	return this.GetFieldData().GetInt64(DriveFileAncestorFieldDepth)
}

func (this *DriveFileAncestor) SetDepth(v *int64) {
	this.GetFieldData().SetInt64(DriveFileAncestorFieldDepth, v)
}

func (this DriveFileAncestor) GetAncestorFile() *DriveFile {
	raw := this.GetFieldData().GetAny(DriveFileAncestorEdgeAncestorFile)
	if raw == nil {
		return nil
	}
	fields, ok := raw.(dmodel.DynamicFields)
	if !ok {
		return nil
	}
	return NewDriveFileFrom(fields)
}
