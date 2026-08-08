package basemodel

import (
	_ "embed"
	"errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)


// The JSON twins of the builders in base_schema.go. The Go builders are kept as-is; these
// exist so that JSON models can reference the same base schemas from "extend_before" /
// "extend_after", which resolve by name through the builder registry.
//
// Fields whose options cannot be expressed in JSON (ServiceInjected) are decorated in Go
// after parsing, via ModelSchemaBuilder.GetField.

//go:embed base_model.json
var baseModelJson string

//go:embed org_base_model.json
var orgBaseModelJson string

//go:embed archivable_model.json
var archivableModelJson string

//go:embed auditable_model.json
var auditableModelJson string

//go:embed auditable_readonly_model.json
var auditableReadonlyModelJson string

//go:embed traceable_model.json
var traceableModelJson string

//go:embed traceable_readonly_model.json
var traceableReadonlyModelJson string

//go:embed versioned_model.json
var versionedModelJson string

func BaseModelSchemaBuilderJson() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(baseModelJson).ExtendBase()
}

func OrgIdModelSchemaBuilderJson() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgBaseModelJson)
}

func ArchivableModelSchemaBuilderJson() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(archivableModelJson)
}

func AuditableModelSchemaBuilderJson() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(auditableModelJson)
}

func AuditableReadonlyModelSchemaBuilderJson() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(auditableReadonlyModelJson)
}

// TraceableModelSchemaBuilderJson parses the field shape from JSON, then attaches the
// service-injection functions, which have no JSON representation.
func TraceableModelSchemaBuilderJson() *dmodel.ModelSchemaBuilder {
	builder := dmodel.ParseModelJson(traceableModelJson)
	builder.GetField(FieldCreatedBy).ServiceInjected(injectCreatedBy)
	builder.GetField(FieldUpdatedBy).ServiceInjected(injectUpdatedBy)

	return builder
}

func TraceableReadonlyModelSchemaBuilderJson() *dmodel.ModelSchemaBuilder {
	builder := dmodel.ParseModelJson(traceableReadonlyModelJson)
	builder.GetField(FieldCreatedBy).ServiceInjected(injectCreatedBy)

	return builder
}

func VersionedModelSchemaBuilderJson() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(versionedModelJson)
}

// RegisterJsonBaseSchemas makes the base schemas resolvable by name from a model JSON's
// "extend_before" / "extend_after". It must run before any module registers a JSON model
// that extends one of them.
func RegisterJsonBaseSchemas() error {
	factories := map[string]func() *dmodel.ModelSchemaBuilder{
		BaseModelSchemaName:              BaseModelSchemaBuilderJson,
		OrgBaseModelSchemaName:           OrgIdModelSchemaBuilderJson,
		ArchivableModelSchemaName:        ArchivableModelSchemaBuilderJson,
		AuditableModelSchemaName:         AuditableModelSchemaBuilderJson,
		AuditableReadonlyModelSchemaName: AuditableReadonlyModelSchemaBuilderJson,
		TraceableModelSchemaName:         TraceableModelSchemaBuilderJson,
		TraceableReadonlyModelSchemaName: TraceableReadonlyModelSchemaBuilderJson,
		VersionedModelSchemaName:         VersionedModelSchemaBuilderJson,
	}

	var err error
	for name, factory := range factories {
		err = errors.Join(err, dmodel.RegisterSchemaBuilderFn(name, factory))
	}

	return err
}
