package model

import (
	"strings"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
)

type ModelSchemaBuilder struct {
	schema        ModelSchema
	shouldBuildDb bool
}

func DefineModel(name string) *ModelSchemaBuilder {
	builder := &ModelSchemaBuilder{
		schema: ModelSchema{
			fields: make(map[string]*ModelField),
		},
		shouldBuildDb: false,
	}
	builder.Name(name)
	return builder
}

func (this *ModelSchemaBuilder) Label(label model.LangJson) *ModelSchemaBuilder {
	this.schema.label = label
	return this
}

func (this *ModelSchemaBuilder) LabelRef(key string) *ModelSchemaBuilder {
	return this.Label(model.LangJson{"$s": key})
}

func (this *ModelSchemaBuilder) Description(description model.LangJson) *ModelSchemaBuilder {
	this.schema.description = description
	return this
}

func (this *ModelSchemaBuilder) Name(name string) *ModelSchemaBuilder {
	this.schema.name = name
	return this
}

func (this *ModelSchemaBuilder) ShouldBuildDb() *ModelSchemaBuilder {
	this.shouldBuildDb = true
	return this
}

func (this *ModelSchemaBuilder) Field(fieldBuilder *FieldBuilder) *ModelSchemaBuilder {
	if fieldBuilder == nil {
		return this
	}
	field := fieldBuilder.Build()
	this.addField(field)

	if field.relation != nil {
		rel := field.relation
		rel.SrcField = field.name
		this.schema.relations = append(this.schema.relations, *rel)
		field.relation = nil
		if rel.Edge != "" {
			this.addImplicitEdgeField(rel)
		}
	}

	return this
}

func (this *ModelSchemaBuilder) CopyField(schema *ModelSchema, fieldName string) *ModelSchemaBuilder {
	newField := copyField(schema, fieldName).Build()
	this.addField(newField)
	return this
}

func (this *ModelSchemaBuilder) CopyFieldN(schemaName string, fieldName string) *ModelSchemaBuilder {
	newField := copyFieldN(schemaName, fieldName).Build()
	this.addField(newField)
	return this
}

func (this *ModelSchemaBuilder) addField(field *ModelField) {
	if err := validateFieldName(field); err != nil {
		panic(errors.Wrapf(err, "addField: model '%s'", this.schema.name))
	}
	if err := validateFieldKeyFlags(field); err != nil {
		panic(errors.Wrapf(err, "addField: model '%s'", this.schema.name))
	}
	if err := validateSingleTenantKey(this.schema.fields, field); err != nil {
		panic(errors.Wrapf(err, "addField: model '%s'", this.schema.name))
	}
	if err := validateNoDuplicateColumn(this.schema.fields, field); err != nil {
		panic(errors.Wrapf(err, "addField: model '%s'", this.schema.name))
	}
	if this.schema.fields == nil {
		this.schema.fields = make(map[string]*ModelField)
	}
	this.schema.fields[field.name] = field
	this.schema.fieldsOrder = append(this.schema.fieldsOrder, field.name)
}

func (this *ModelSchemaBuilder) Extend(builder *ModelSchemaBuilder) *ModelSchemaBuilder {
	for _, fieldName := range builder.schema.fieldsOrder {
		this.addField(builder.schema.fields[fieldName])
	}
	this.schema.relations = append(this.schema.relations, builder.schema.relations...)
	this.schema.compositeUniques = append(this.schema.compositeUniques, builder.schema.compositeUniques...)
	return this
}

func (this *ModelSchemaBuilder) addImplicitEdgeField(rel *ModelRelation) {
	isArray := rel.RelationType == RelationTypeOneToMany || rel.RelationType == RelationTypeManyToMany
	dataType := FieldDataType(FieldDataTypeModel())
	if isArray {
		dataType = dataType.ArrayType()
	}
	this.Field(DefineField().Name(rel.Edge).DataType(dataType))
}

func (this *ModelSchemaBuilder) TableName(tableName string) *ModelSchemaBuilder {
	this.schema.tableName = tableName
	return this
}

func (this *ModelSchemaBuilder) CompositeUnique(composite ...string) *ModelSchemaBuilder {
	if len(composite) > 0 {
		this.schema.compositeUniques = append(this.schema.compositeUniques, composite)
	}
	return this
}

func (this *ModelSchemaBuilder) SetCompositeUniques(allUniques [][]string) *ModelSchemaBuilder {
	this.schema.compositeUniques = allUniques
	return this
}

func (this *ModelSchemaBuilder) Build() *ModelSchema {
	schema := &this.schema
	if this.shouldBuildDb {
		ft.PanicOnErr(populateDbMetadata(schema))
	}
	return schema
}

func copyField(schema *ModelSchema, fieldName string) *FieldBuilder {
	field := schema.MustField(fieldName)
	copiedField := field.Copy()
	return &FieldBuilder{
		field: copiedField,
	}
}

func copyFieldN(schemaName string, fieldName string) *FieldBuilder {
	field := schemaRegistry.Field(schemaName, fieldName)
	copiedField := field.Copy()
	return &FieldBuilder{
		field: copiedField,
	}
}

func validateFieldName(field *ModelField) error {
	if field.name == "" {
		return errors.Errorf("validateFieldName: field name is required")
	}
	return nil
}

func validateFieldKeyFlags(field *ModelField) error {
	if field.isPrimaryKey && field.isTenantKey {
		return errors.Errorf(
			"validateFieldKeyFlags: field '%s': isPrimaryKey and isTenantKey are mutually exclusive", field.name)
	}
	return nil
}

func validateSingleTenantKey(existingFields map[string]*ModelField, newField *ModelField) error {
	if !newField.isTenantKey {
		return nil
	}
	for _, f := range existingFields {
		if f != nil && f.isTenantKey {
			return errors.Errorf(
				"validateSingleTenantKey: field '%s' cannot be tenant key: '%s' is already the tenant key",
				newField.name, f.name)
		}
	}
	return nil
}

func validateNoDuplicateColumn(existingFields map[string]*ModelField, newField *ModelField) error {
	columnName := newField.name
	for _, f := range existingFields {
		if f != nil && f.name == columnName {
			return errors.Errorf("validateNoDuplicateColumn: duplicate column '%s'", columnName)
		}
	}
	return nil
}

type FieldBuilder struct {
	field *ModelField
}

func DefineField() *FieldBuilder {
	return &FieldBuilder{
		field: &ModelField{},
	}
}

func (this *FieldBuilder) Description(description model.LangJson) *FieldBuilder {
	this.field.description = description
	return this
}

func (this *FieldBuilder) DataType(dataType FieldDataType) *FieldBuilder {
	this.field.dataType = dataType
	return this
}

func (this *FieldBuilder) Foreign(relationBuilder *RelationBuilder) *FieldBuilder {
	this.field.relation = relationBuilder.Build()
	return this
}

func (this *FieldBuilder) Label(label model.LangJson) *FieldBuilder {
	this.field.label = label
	return this
}

func (this *FieldBuilder) LabelRef(key string) *FieldBuilder {
	return this.Label(model.LangJson{model.LabelRefLanguageCode: key})
}

func (this *FieldBuilder) Name(name string) *FieldBuilder {
	this.field.name = strings.TrimSpace(name)
	return this
}

// Indicates that the field value cannot be set by user but by the system.
// Any input value will be silently ignored when creating or updating the model.
// If a default value is registered, it will be used in create operations.
func (this *FieldBuilder) ReadOnly() *FieldBuilder {
	this.field.isReadOnly = true
	return this
}

// Uses this for schemas which is used for validation and not for SQL generation.
func (this *FieldBuilder) Required() *FieldBuilder {
	this.field.isRequiredForCreate = true
	this.field.isRequiredForUpdate = true
	return this
}

// Causes the field to be required for create operations,
// and determines the "NOT NULL" constraint for the database column.
// Missing field error will occur when the input value is nil and the field doesn't have a registered default value.
func (this *FieldBuilder) RequiredForCreate() *FieldBuilder {
	this.field.isRequiredForCreate = true
	return this
}

// Causes the field to be required for update operations,
// but doesn't affect the generated CREATE SQL query.
// Missing field error will occur when the input value is nil REGARDLESS the field has a registered default value or not.
func (this *FieldBuilder) RequiredForUpdate() *FieldBuilder {
	this.field.isRequiredForUpdate = true
	return this
}

func (this *FieldBuilder) IsRequired(isRequired bool) *FieldBuilder {
	this.field.isRequiredForCreate = isRequired
	this.field.isRequiredForUpdate = isRequired
	return this
}

func (this *FieldBuilder) IsRequiredForCreate(isRequired bool) *FieldBuilder {
	this.field.isRequiredForCreate = isRequired
	return this
}

func (this *FieldBuilder) IsRequiredForUpdate(isRequired bool) *FieldBuilder {
	this.field.isRequiredForUpdate = isRequired
	return this
}

func (this *FieldBuilder) IsReadOnly(isReadOnly bool) *FieldBuilder {
	this.field.isReadOnly = isReadOnly
	return this
}

func (this *FieldBuilder) PrimaryKey() *FieldBuilder {
	this.field.isPrimaryKey = true
	this.RequiredForCreate() // NOT NULL column
	this.RequiredForUpdate()
	this.ReadOnly()
	return this
}

func (this *FieldBuilder) TenantKey() *FieldBuilder {
	this.field.isTenantKey = true
	return this
}

func (this *FieldBuilder) Rule(rule FieldRule) *FieldBuilder {
	rules := this.field.rules
	rules = append(rules, &rule)
	this.field.rules = rules
	return this
}

// Sets the default value for the field.
// Default value is only used for create operations and when the input field is nil.
// Read-only fields are always set to the default value regardless of the input.
// The precedence is: Default > DefaultFn > UseTypeDefault.
func (this *FieldBuilder) Default(val any) *FieldBuilder {
	this.field.defaultValue = util.ToPtr(Value(val))
	return this
}

// Registers a function to generate the default value for the field.
// Default value is only used for create operations and when the input field is nil.
// Read-only fields are always set to the default value regardless of the input.
// The precedence is: Default > DefaultFn > UseTypeDefault.
func (this *FieldBuilder) DefaultFn(fn func() any) *FieldBuilder {
	this.field.defaultFn = fn
	return this
}

// Indicates that the field should use the default value from the type definition.
// Default value is only used for create operations and when the input field is nil.
// Read-only fields are always set to the default value regardless of the input.
// The precedence is: Default > DefaultFn > UseTypeDefault.
func (this *FieldBuilder) UseTypeDefault() *FieldBuilder {
	this.field.useTypeDefault = true
	return this
}

func (this *FieldBuilder) SetUseTypeDefault(useTypeDefault bool) *FieldBuilder {
	this.field.useTypeDefault = useTypeDefault
	return this
}

func (this *FieldBuilder) Unique() *FieldBuilder {
	this.field.isUnique = true
	return this
}

// Indicates that the field value is used for versioning the model,
// which means it is both read-only and required for update operations.
func (this *FieldBuilder) VersioningKey() *FieldBuilder {
	this.field.isVersioningKey = true
	this.RequiredForCreate() // NOT NULL column
	this.RequiredForUpdate()
	this.ReadOnly()
	return this
}

func (this *FieldBuilder) Build() *ModelField {
	if this.field.name == "" {
		panic("field name is required")
	}
	return this.field
}

type RelationBuilder struct {
	relation *ModelRelation
}

func Edge(edgeName string) *RelationBuilder {
	return &RelationBuilder{
		relation: &ModelRelation{
			Edge: edgeName,
		},
	}
}

func (this *RelationBuilder) Label(label model.LangJson) *RelationBuilder {
	this.relation.label = label
	return this
}

func (this *RelationBuilder) OneToOne(schemaName string, destField string) *RelationBuilder {
	this.relation.RelationType = RelationTypeOneToOne
	this.relation.DestSchemaName = schemaName
	this.relation.DestField = destField
	return this
}

func (this *RelationBuilder) OneToMany(schemaName string, destField string) *RelationBuilder {
	this.relation.RelationType = RelationTypeOneToMany
	this.relation.DestSchemaName = schemaName
	this.relation.DestField = destField
	return this
}

func (this *RelationBuilder) ManyToOne(schemaName string, targetField string) *RelationBuilder {
	this.relation.RelationType = RelationTypeManyToOne
	this.relation.DestSchemaName = schemaName
	this.relation.DestField = targetField
	return this
}

func (this *RelationBuilder) ManyToMany(throughTableName string, throughSrcCol string, throughDestCol string) *RelationBuilder {
	this.relation.RelationType = RelationTypeManyToMany
	this.relation.ThroughTableName = throughTableName
	this.relation.ThroughSrcCol = throughSrcCol
	this.relation.ThroughDestCol = throughDestCol
	return this
}

func (this *RelationBuilder) OnDelete(onDelete RelationCascade) *RelationBuilder {
	this.relation.OnDelete = onDelete
	return this
}

func (this *RelationBuilder) OnUpdate(onUpdate RelationCascade) *RelationBuilder {
	this.relation.OnUpdate = onUpdate
	return this
}

func (this *RelationBuilder) Build() *ModelRelation {
	return this.relation
}
