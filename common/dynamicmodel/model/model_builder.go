package model

import (
	"regexp"
	"strings"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
)

var indexNameRegex = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

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
	this.schema.toRelations = append(this.schema.toRelations, builder.schema.toRelations...)
	this.schema.fromRelations = append(this.schema.fromRelations, builder.schema.fromRelations...)
	this.schema.compositeUniques = append(this.schema.compositeUniques, builder.schema.compositeUniques...)
	this.schema.partialUniqueGroups = append(this.schema.partialUniqueGroups, builder.schema.partialUniqueGroups...)
	this.schema.searchIndexGroups = append(this.schema.searchIndexGroups, builder.schema.searchIndexGroups...)
	this.schema.exclusiveFieldGroups = append(
		this.schema.exclusiveFieldGroups, builder.schema.exclusiveFieldGroups...)
	return this
}

func (this *ModelSchemaBuilder) ExtendBase() *ModelSchemaBuilder {
	if baseBuilder != nil {
		this.Extend(baseBuilder)
	}
	return this
}

// ExclusiveFields registers one exclusive group: exactly one of the listed fields must be
// non-empty on validate. The slice may contain any number of field names (minimum two). Call
// multiple times to register multiple independent groups. Each name must exist on the schema
// when Build runs.
func (this *ModelSchemaBuilder) ExclusiveFields(fieldNames ...string) *ModelSchemaBuilder {
	if len(fieldNames) < 2 {
		panic(errors.New("ExclusiveFieldGroup: at least two field names are required"))
	}
	group := append([]string{}, fieldNames...)
	this.schema.exclusiveFieldGroups = append(this.schema.exclusiveFieldGroups, group)
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

func (this *ModelSchemaBuilder) EdgeFrom(rb *RelationBuilder) *ModelSchemaBuilder {
	rel := rb.Build()
	if rel.InversePeerSchemaName == "" || rel.InversePeerEdgeName == "" {
		panic(errors.New("EdgeFrom: Existing(srcSchemaName, srcEdgeName) is required"))
	}
	if rel.RelationType != "" {
		panic(errors.New("EdgeFrom: do not set relation type; it is derived from the peer EdgeTo"))
	}
	if rel.Edge == "" {
		panic(errors.New("EdgeFrom: edge name is required"))
	}
	this.schema.fromRelations = append(this.schema.fromRelations, *rel)
	return this
}

func (this *ModelSchemaBuilder) EdgeTo(rb *RelationBuilder) *ModelSchemaBuilder {
	rel := rb.Build()
	if rel.RelationType == RelationTypeManyToMany {
		this.validateManyToManyCascade(*rel)
		// Will be set by SchemaRegistry.FinalizeRelations()
		rel.SrcField = ""
		rel.DestField = ""
	}
	this.schema.toRelations = append(this.schema.toRelations, *rel)
	if rel.Edge != "" {
		this.addImplicitEdgeField(rel)
	}
	return this
}

func (this *ModelSchemaBuilder) validateManyToManyCascade(rel ModelRelation) {
	if rel.OnDelete != "" && rel.OnDelete != RelationCascadeNoAction &&
		rel.OnDelete != RelationCascadeCascade {
		panic(errors.Errorf(
			"validateManyToManyCascade: relation '%s': OnDelete must be NO ACTION or CASCADE", rel.Edge))
	}
	if rel.OnUpdate != "" && rel.OnUpdate != RelationCascadeNoAction &&
		rel.OnUpdate != RelationCascadeCascade {
		panic(errors.Errorf(
			"validateManyToManyCascade: relation '%s': OnUpdate must be NO ACTION or CASCADE", rel.Edge))
	}
}

func (this *ModelSchemaBuilder) CompositeUnique(composite ...string) *ModelSchemaBuilder {
	if len(composite) > 0 {
		this.schema.compositeUniques = append(this.schema.compositeUniques, composite)
	}
	return this
}

// PartialUnique registers a partial unique index on two columns: exactly one must be requiredForCreate
// (NOT NULL) and the other nullable. Enforced in Build() when ShouldBuildDb is set.
func (this *ModelSchemaBuilder) PartialUnique(notNullField, nullableField string) *ModelSchemaBuilder {
	a := strings.TrimSpace(notNullField)
	b := strings.TrimSpace(nullableField)
	if a != "" && b != "" {
		this.PartialUniqueGroup(PartialUniqueGroupParam{
			NotNullFields: []string{a},
			NullableField: b,
		})
	}
	return this
}

// SearchIndex causes the migration script to generate CREATE INDEX statement for the given fields.
// Field order matters: Place the most frequently queried column or
// the one with the highest selectivity (most unique values) first.
func (this *ModelSchemaBuilder) SearchIndex(fields ...string) *ModelSchemaBuilder {
	return this.SearchIndexGroup(SearchIndexGroupParam{Fields: fields})
}

var baseBuilder *ModelSchemaBuilder

func SetBaseModelSchemaBuilder(builder *ModelSchemaBuilder) {
	baseBuilder = builder
}

type SearchIndexGroupParam struct {
	// If not specified, a default name will be generated from all field names.
	// Recommend to provide an index name when the number of fields is more than 2.
	IndexName string
	// Field order matters: Place the most frequently queried column or
	// the one with the highest selectivity (most unique values) first.
	Fields []string
}

// SearchIndexGroup causes the migration script to generate CREATE INDEX statement for the given fields.
func (this *ModelSchemaBuilder) SearchIndexGroup(group SearchIndexGroupParam) *ModelSchemaBuilder {
	if len(group.Fields) == 0 {
		panic(errors.New("SearchIndexGroup: field list must not be empty"))
	}
	fields := array.Map(group.Fields, func(fieldName string) string {
		trimName := strings.TrimSpace(fieldName)
		if trimName == "" {
			panic(errors.Errorf("SearchIndexGroup: field name must not be empty: %s", fieldName))
		}
		return trimName
	})

	this.schema.searchIndexGroups = append(this.schema.searchIndexGroups, SearchIndexGroupParam{
		IndexName: mustNormalizeIndexName(group.IndexName),
		Fields:    fields,
	})
	return this
}

type PartialUniqueGroupParam struct {
	IndexName     string
	NotNullFields []string
	NullableField string
}

func (this *ModelSchemaBuilder) PartialUniqueGroup(group PartialUniqueGroupParam) *ModelSchemaBuilder {
	indexName := mustNormalizeIndexName(group.IndexName)
	nullableField := strings.TrimSpace(group.NullableField)
	notNullFields := array.Map(group.NotNullFields, func(fieldName string) string {
		trimName := strings.TrimSpace(fieldName)
		if trimName == "" {
			panic(errors.Errorf("PartialUniqueGroup: field name must not be empty: %s", fieldName))
		}
		return trimName
	})
	this.schema.partialUniqueGroups = append(this.schema.partialUniqueGroups, PartialUniqueGroupParam{
		IndexName:     indexName,
		NotNullFields: notNullFields,
		NullableField: nullableField,
	})
	return this
}

func mustNormalizeIndexName(raw string) string {
	indexName := strings.TrimSpace(raw)
	if indexName == "" {
		return ""
	}
	if !indexNameRegex.MatchString(indexName) {
		panic(errors.Errorf(
			"mustNormalizeIndexName: invalid index name '%s'; only alphanumeric and '_' are allowed",
			indexName))
	}
	if !strings.HasSuffix(indexName, "_idx") {
		indexName += "_idx"
	}
	return indexName
}

func (this *ModelSchemaBuilder) SetCompositeUniques(allUniques [][]string) *ModelSchemaBuilder {
	this.schema.compositeUniques = allUniques
	return this
}

func (this *ModelSchemaBuilder) Build() *ModelSchema {
	schema := &this.schema
	if err := validateExclusiveFieldGroups(schema); err != nil {
		panic(errors.Wrap(err, "Build"))
	}
	if err := validateRequiredWithFields(schema); err != nil {
		panic(errors.Wrap(err, "Build"))
	}
	if this.shouldBuildDb {
		ft.PanicOnErr(populateDbMetadata(schema))
	}
	return schema
}

func validateExclusiveFieldGroups(schema *ModelSchema) error {
	for gi, group := range schema.exclusiveFieldGroups {
		if len(group) < 2 {
			return errors.Errorf("exclusive field group %d: at least two field names required", gi)
		}
		for _, name := range group {
			if _, ok := schema.Field(name); !ok {
				return errors.Errorf(
					"exclusive field group %d: field %q is not defined on schema %q",
					gi, name, schema.name)
			}
		}
	}
	return nil
}

func validateRequiredWithFields(schema *ModelSchema) error {
	for _, field := range schema.fields {
		if field.requiredWithFieldName != "" {
			if _, ok := schema.Field(field.requiredWithFieldName); !ok {
				return errors.Errorf("validateRequiredWithFields: field '%s' depends on undefined field '%s'", field.name, field.requiredWithFieldName)
			}
		}
	}
	return nil
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
func (this *FieldBuilder) AutoGenerated() *FieldBuilder {
	this.field.isAutoGenerated = true
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

func (this *FieldBuilder) RequiredWith(otherFieldName string) *FieldBuilder {
	this.field.requiredWithFieldName = otherFieldName
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

func (this *FieldBuilder) IsAutoGenerated(isAutoGenerated bool) *FieldBuilder {
	this.field.isAutoGenerated = isAutoGenerated
	return this
}

// Allows setting value on create but not on update.
func (this *FieldBuilder) NoUpdate() *FieldBuilder {
	this.field.noUpdate = true
	return this
}

func (this *FieldBuilder) PrimaryKey(isAutoGenerated ...bool) *FieldBuilder {
	this.field.isPrimaryKey = true
	this.RequiredForCreate() // NOT NULL column
	this.RequiredForUpdate()
	if len(isAutoGenerated) == 0 {
		this.IsAutoGenerated(true)
	} else {
		this.IsAutoGenerated(isAutoGenerated[0])
	}
	return this
}

func (this *FieldBuilder) TenantKey() *FieldBuilder {
	this.field.isTenantKey = true
	this.RequiredForCreate() // NOT NULL column
	this.IsAutoGenerated(true)
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
	this.AutoGenerated()
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

func (this *RelationBuilder) OneToOne(destSchemaName string, srcDestKeyMap DynamicFields) *RelationBuilder {
	this.relation.RelationType = RelationTypeOneToOne
	this.relation.DestSchemaName = strings.TrimSpace(destSchemaName)
	this.relation.UnvalidatedFkMap = srcDestKeyMap
	return this
}

func (this *RelationBuilder) OneToMany(destSchemaName string, srcDestKeyMap DynamicFields) *RelationBuilder {
	this.relation.RelationType = RelationTypeOneToMany
	this.relation.DestSchemaName = strings.TrimSpace(destSchemaName)
	this.relation.UnvalidatedFkMap = srcDestKeyMap
	return this
}

func (this *RelationBuilder) ManyToOne(destSchemaName string, srcDestKeyMap DynamicFields) *RelationBuilder {
	this.relation.RelationType = RelationTypeManyToOne
	this.relation.DestSchemaName = strings.TrimSpace(destSchemaName)
	this.relation.UnvalidatedFkMap = srcDestKeyMap
	return this
}

func (this *RelationBuilder) Existing(srcSchemaName, srcEdgeName string) *RelationBuilder {
	this.relation.InversePeerSchemaName = strings.TrimSpace(srcSchemaName)
	this.relation.InversePeerEdgeName = strings.TrimSpace(srcEdgeName)
	return this
}

func (this *RelationBuilder) ManyToMany(peerSchemaName, throughSchemaName, srcFieldPrefix string) *RelationBuilder {
	this.relation.RelationType = RelationTypeManyToMany
	this.relation.DestSchemaName = peerSchemaName
	this.relation.M2mThroughSchemaName = throughSchemaName
	this.relation.M2mSrcFieldPrefix = srcFieldPrefix
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
	if this.relation.OnDelete == "" {
		this.relation.OnDelete = RelationCascadeNoAction
	}
	if this.relation.OnUpdate == "" {
		this.relation.OnUpdate = RelationCascadeNoAction
	}
	return this.relation
}
