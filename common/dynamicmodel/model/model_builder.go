package model

import (
	"context"
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
	builder.Label(model.NewLangJsonRefSf("%s.label", name))
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
	this.schema.partialUniques = append(this.schema.partialUniques, builder.schema.partialUniques...)
	this.schema.searchIndexGroups = append(this.schema.searchIndexGroups, builder.schema.searchIndexGroups...)
	this.schema.exclusiveRequiredFieldGroups = append(
		this.schema.exclusiveRequiredFieldGroups, builder.schema.exclusiveRequiredFieldGroups...)
	// Inherited only when this schema has not declared its own, so a concrete model always wins
	// over whatever a base builder happens to name.
	if this.schema.recordLabelField == "" {
		this.schema.recordLabelField = builder.schema.recordLabelField
	}
	if this.schema.recordSubLabelField == "" {
		this.schema.recordSubLabelField = builder.schema.recordSubLabelField
	}
	return this
}

// ExtendByName extends this schema with the builder registered under schemaName
// via RegisterSchemaBuilderFn. Panics when no builder is registered, because a JSON
// model referencing an unknown base schema is a programming error, not user input.
func (this *ModelSchemaBuilder) ExtendByName(schemaName string) *ModelSchemaBuilder {
	builder := GetSchemaBuilder(schemaName)
	if builder == nil {
		panic(errors.Errorf(
			"ExtendByName: model '%s' extends '%s' which is not registered; "+
				"register it with RegisterSchemaBuilderFn before parsing this model",
			this.schema.name, schemaName,
		))
	}
	return this.Extend(builder)
}

// HasField reports whether a field of the given name has been added to this schema.
func (this *ModelSchemaBuilder) HasField(name string) bool {
	_, exists := this.schema.fields[name]
	return exists
}

// GetField returns a FieldBuilder wrapping an already-added field, so callers can apply
// options that cannot be expressed in JSON (ServiceInjected, DefaultFn). The returned
// builder mutates the field in place, so there is no need to re-add it.
// Panics when the field does not exist.
func (this *ModelSchemaBuilder) GetField(name string) *FieldBuilder {
	field, exists := this.schema.fields[name]
	if !exists {
		panic(errors.Errorf("GetField: field '%s' not found in model '%s'", name, this.schema.name))
	}
	return &FieldBuilder{field: field}
}

func (this *ModelSchemaBuilder) ExtendBase() *ModelSchemaBuilder {
	if baseBuilder != nil {
		this.Extend(baseBuilder)
	}
	return this
}

// ExclusiveRequiredFields registers one exclusive group: exactly one of the listed fields must be
// non-empty and required on validate. The slice may contain any number of field names (minimum two). Call
// multiple times to register multiple independent groups. Each name must exist on the schema
// when Build runs.
func (this *ModelSchemaBuilder) ExclusiveRequiredFields(fieldNames ...string) *ModelSchemaBuilder {
	if len(fieldNames) < 2 {
		panic(errors.New("ExclusiveRequiredFields: at least two field names are required"))
	}
	group := append([]string{}, fieldNames...)
	this.schema.exclusiveRequiredFieldGroups = append(this.schema.exclusiveRequiredFieldGroups, group)
	return this
}

// RecordLabelField declares the field that identifies a record of this model to a human — the text
// a client shows wherever a record stands in for itself, such as a relation picker or a breadcrumb.
// The named field must exist on the schema when Build runs.
func (this *ModelSchemaBuilder) RecordLabelField(fieldName string) *ModelSchemaBuilder {
	this.schema.recordLabelField = fieldName
	return this
}

// RecordSubLabelField declares an optional secondary field, shown beneath the main label to tell
// apart records that share one. The named field must exist on the schema when Build runs.
func (this *ModelSchemaBuilder) RecordSubLabelField(fieldName string) *ModelSchemaBuilder {
	this.schema.recordSubLabelField = fieldName
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

type CompositeUniqueParam struct {
	// IndexName is optional. When empty, the constraint name is derived as
	// "{tableName}_{tenantKey}_{Fields...}". When set, it replaces that whole stem, so it must
	// carry the table prefix itself. The "_ukey" suffix is always appended by the query builder;
	// never write it here. See docs/wiki "04. Dynamic schema" for the 63-byte naming rules.
	IndexName string
	// Fields must all be requiredForCreate. Use PartialUnique when one of them is nullable.
	Fields []string
}

// CompositeUnique registers a multi-column UNIQUE constraint. All columns must be requiredForCreate,
// enforced in Build() when ShouldBuildDb is set.
func (this *ModelSchemaBuilder) CompositeUnique(param CompositeUniqueParam) *ModelSchemaBuilder {
	if len(param.Fields) == 0 {
		panic(errors.New("CompositeUnique: field list must not be empty"))
	}
	fields := array.Map(param.Fields, func(fieldName string) string {
		trimName := strings.TrimSpace(fieldName)
		if trimName == "" {
			panic(errors.Errorf("CompositeUnique: field name must not be empty: %s", fieldName))
		}
		return trimName
	})
	this.schema.compositeUniques = append(this.schema.compositeUniques, CompositeUniqueParam{
		IndexName: mustValidateIndexName(param.IndexName),
		Fields:    fields,
	})
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
		IndexName: mustNormalizeSearchIndexName(group.IndexName),
		Fields:    fields,
	})
	return this
}

type PartialUniqueParam struct {
	// IndexName is optional. When empty, the index name is derived as
	// "{tableName}_{tenantKey}_{NotNullFields...}_{NullableField}". When set, it replaces that whole
	// stem, so it must carry the table prefix itself. The "_ukey_notnull" and "_ukey_null" suffixes
	// are always appended by the query builder; never write them here.
	IndexName string
	// NotNullFields must all be requiredForCreate.
	NotNullFields []string
	// NullableField must NOT be requiredForCreate.
	NullableField string
}

// PartialUnique registers a pair of partial unique indexes: one over NotNullFields plus
// NullableField where the latter IS NOT NULL, and one over NotNullFields alone where it IS NULL.
// This is how tenant/org-scoped uniqueness is expressed.
// Enforced in Build() when ShouldBuildDb is set.
func (this *ModelSchemaBuilder) PartialUnique(param PartialUniqueParam) *ModelSchemaBuilder {
	indexName := mustValidateIndexName(param.IndexName)
	nullableField := strings.TrimSpace(param.NullableField)
	notNullFields := array.Map(param.NotNullFields, func(fieldName string) string {
		trimName := strings.TrimSpace(fieldName)
		if trimName == "" {
			panic(errors.Errorf("PartialUnique: field name must not be empty: %s", fieldName))
		}
		return trimName
	})
	this.schema.partialUniques = append(this.schema.partialUniques, PartialUniqueParam{
		IndexName:     indexName,
		NotNullFields: notNullFields,
		NullableField: nullableField,
	})
	return this
}

// mustValidateIndexName checks the character set of a caller-provided index name and returns it
// unchanged. Unique index names must NOT gain an "_idx" suffix: the query builder appends
// "_ukey", "_ukey_notnull" or "_ukey_null" to them.
func mustValidateIndexName(raw string) string {
	indexName := strings.TrimSpace(raw)
	if indexName == "" {
		return ""
	}
	if !indexNameRegex.MatchString(indexName) {
		panic(errors.Errorf(
			"mustValidateIndexName: invalid index name '%s'; only alphanumeric and '_' are allowed",
			indexName))
	}
	return indexName
}

// mustNormalizeSearchIndexName validates a non-unique index name and appends the "_idx" suffix
// convention when the caller did not.
func mustNormalizeSearchIndexName(raw string) string {
	indexName := mustValidateIndexName(raw)
	if indexName == "" {
		return ""
	}
	if !strings.HasSuffix(indexName, "_idx") {
		indexName += "_idx"
	}
	return indexName
}

func (this *ModelSchemaBuilder) Build() *ModelSchema {
	schema := &this.schema
	if err := validateExclusiveFieldGroups(schema); err != nil {
		panic(errors.Wrap(err, "Build"))
	}
	if err := validateRequiredWithFields(schema); err != nil {
		panic(errors.Wrap(err, "Build"))
	}
	if err := validateRecordLabelFields(schema); err != nil {
		panic(errors.Wrap(err, "Build"))
	}
	// Runs here rather than inside populateDbMetadata, which only runs when shouldBuildDb is set:
	// a validation-only schema must reject the same contradictions.
	if err := validateVirtualFields(schema); err != nil {
		panic(errors.Wrap(err, "Build"))
	}
	// A computed field is virtual, so validateVirtualFields above already rejects every
	// role-contradiction (key, required, indexed, ...) for it; only computed-specific rules run here.
	if err := validateComputedFields(schema); err != nil {
		panic(errors.Wrap(err, "Build"))
	}
	if this.shouldBuildDb {
		ft.PanicOnErr(populateDbMetadata(schema))
	}
	return schema
}

// validateVirtualFields rejects a virtual field that also claims a role requiring a column.
// Each combination below would otherwise fail silently and far from its cause: the field is
// dropped from every write, so a "required" or "unique" virtual field is a promise nothing can
// keep.
func validateVirtualFields(schema *ModelSchema) error {
	for _, name := range schema.fieldsOrder {
		field, ok := schema.fields[name]
		if !ok || field == nil || !field.IsVirtual() {
			continue
		}
		if field.IsVirtualModelField() {
			return errors.Errorf(
				"field %q: Virtual() is for scalar fields; a model-typed field is already virtual",
				name)
		}
		if field.IsPrimaryKey() || field.IsTenantKey() || field.IsVersioningKey() || field.IsUnique() {
			return errors.Errorf(
				"field %q: a virtual field has no column and cannot be a primary, tenant or "+
					"versioning key, nor unique", name)
		}
		if field.IsRequiredForCreate() || field.IsRequiredForUpdate() || field.requiredWithFieldName != "" {
			return errors.Errorf("field %q: a virtual field is never written and cannot be required", name)
		}
		if field.IsAutoGenerated() || field.IsServiceInjected() {
			return errors.Errorf(
				"field %q: a virtual field is filled after the read; auto-generated and "+
					"service-injected apply to written fields", name)
		}
		if field.IsNoUpdate() {
			return errors.Errorf("field %q: NoUpdate is meaningless on a virtual field", name)
		}
		if err := assertVirtualFieldNotIndexed(schema, name); err != nil {
			return err
		}
	}
	return nil
}

// validateComputedFields applies the computed-specific rules on top of the virtual-field rules
// (a computed field is virtual, so validateVirtualFields has already run for it). The expression
// itself is validated at schema finalize time, when the whole registry is available for
// resolving cross-schema references.
func validateComputedFields(schema *ModelSchema) error {
	for _, name := range schema.fieldsOrder {
		field, ok := schema.fields[name]
		if !ok || !field.IsComputed() {
			continue
		}
		if field.ComputedIsStored() {
			return errors.Errorf(
				"field %q: stored computed fields are not yet supported; declare is_stored: false", name)
		}
	}
	return nil
}

// assertVirtualFieldNotIndexed rejects a virtual field named by anything that needs a column to
// point at: an index, a uniqueness constraint, an exclusive group, or a record label.
func assertVirtualFieldNotIndexed(schema *ModelSchema, name string) error {
	groups := map[string][][]string{"exclusive group": schema.exclusiveRequiredFieldGroups}
	for _, param := range schema.compositeUniques {
		groups["composite unique"] = append(groups["composite unique"], param.Fields)
	}
	for _, param := range schema.partialUniques {
		groups["partial unique"] = append(
			groups["partial unique"], param.NotNullFields, []string{param.NullableField})
	}
	for _, param := range schema.searchIndexGroups {
		groups["search index"] = append(groups["search index"], param.Fields)
	}

	for kind, fieldGroups := range groups {
		for _, group := range fieldGroups {
			for _, member := range group {
				if member == name {
					return errors.Errorf("field %q: a virtual field cannot appear in a %s", name, kind)
				}
			}
		}
	}
	if schema.recordLabelField == name || schema.recordSubLabelField == name {
		return errors.Errorf(
			"field %q: a virtual field cannot be the record label; clients read that column directly",
			name)
	}
	return nil
}

func validateExclusiveFieldGroups(schema *ModelSchema) error {
	for gi, group := range schema.exclusiveRequiredFieldGroups {
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

// validateRecordLabelFields checks that a declared record label points at a real field. Declaring
// one is not yet required: most schemas predate the property, so an absent label is legal and
// leaves clients to fall back to the primary key.
func validateRecordLabelFields(schema *ModelSchema) error {
	declared := map[string]string{
		"recordLabelField":    schema.recordLabelField,
		"recordSubLabelField": schema.recordSubLabelField,
	}
	for property, fieldName := range declared {
		if fieldName == "" {
			continue
		}
		if _, ok := schema.Field(fieldName); !ok {
			return errors.Errorf(
				"%s: field %q is not defined on schema %q", property, fieldName, schema.name)
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
	return this.Label(model.LangJson{model.LanguageCodeRef: key})
}

func (this *FieldBuilder) Name(name string) *FieldBuilder {
	this.field.name = strings.TrimSpace(name)
	this.Label(model.NewLangJsonRefSf("fields.%s", name))
	return this
}

// Indicates that the field value cannot be set by user but by the system.
// Any input value will be silently ignored when creating or updating the model.
// If a default value is registered, it will be used in create operations.
func (this *FieldBuilder) AutoGenerated() *FieldBuilder {
	this.field.isAutoGenerated = true
	return this
}

func (this *FieldBuilder) ServiceInjected(injectFn func(ctx context.Context, forEdit bool) any) *FieldBuilder {
	this.field.isServiceInjected = true
	this.field.injectFn = injectFn
	return this
}

// A shortcut to set both RequiredForCreate() and RequiredForUpdate() at once.
// Use this for schemas used for validation and not for SQL generation.
func (this *FieldBuilder) RequiredAlways() *FieldBuilder {
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

// Virtual marks a scalar field that has no database column. It is selectable and returned in
// results, but never written, and never filtered or sorted in SQL unless the owning module
// rewrites it to the edge path it derives from. A service fills it after the read.
func (this *FieldBuilder) Virtual() *FieldBuilder {
	this.field.isVirtual = true
	return this
}

// Computed marks the field's value as the result of a declared calculation, never user input.
// The expression comes from the computed package's chained constructors:
//
//	Computed(false, computed.Sub(computed.F("on_hand_quantity"), computed.F("reserved_quantity")))
//	Computed(false, computed.Related("template.name"))
//
// isStored=false computes the value when the resource is read; the field is virtual (no column).
// isStored=true — compute at write time with source-change propagation — is reserved for a
// future phase and rejected at Build(). The expression is typed `any` to avoid an import cycle
// with the computed package; it is validated and resolved at schema finalize time.
func (this *FieldBuilder) Computed(isStored bool, expression any) *FieldBuilder {
	if isNil(expression) {
		panic(errors.Errorf("field %q: Computed requires an expression", this.field.name))
	}
	this.field.isVirtual = true
	this.field.computedExpr = expression
	this.field.computedIsStored = isStored
	return this
}

func (this *FieldBuilder) Placeholder(placeholder model.LangJson) *FieldBuilder {
	this.field.placeholder = placeholder
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
