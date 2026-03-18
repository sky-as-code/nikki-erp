package schema

import (
	"strings"

	"github.com/sky-as-code/nikki-erp/common/model"
	"go.bryk.io/pkg/errors"
)

type EntitySchemaBuilder struct {
	schema EntitySchema
}

func DefineEntity(name string) *EntitySchemaBuilder {
	builder := &EntitySchemaBuilder{
		schema: EntitySchema{
			fields: make(map[string]*EntityField),
		},
	}
	builder.Name(name)
	return builder
}

func (this *EntitySchemaBuilder) Label(label model.LangJson) *EntitySchemaBuilder {
	this.schema.label = label
	return this
}

func (this *EntitySchemaBuilder) LabelRef(key string) *EntitySchemaBuilder {
	return this.Label(model.LangJson{"$s": key})
}

func (this *EntitySchemaBuilder) Description(description model.LangJson) *EntitySchemaBuilder {
	this.schema.description = description
	return this
}

func (this *EntitySchemaBuilder) Name(name string) *EntitySchemaBuilder {
	this.schema.name = name
	return this
}

func (this *EntitySchemaBuilder) Field(fieldBuilder *FieldBuilder) *EntitySchemaBuilder {
	if fieldBuilder == nil {
		return this
	}

	field := fieldBuilder.Build()
	if err := validateFieldName(field); err != nil {
		panic(errors.Wrapf(err, "entity '%s'", this.schema.name))
	}
	if err := validateFieldKeyFlags(field); err != nil {
		panic(errors.Wrapf(err, "entity '%s'", this.schema.name))
	}
	if err := validateSingleTenantKey(this.schema.fields, field); err != nil {
		panic(errors.Wrapf(err, "entity '%s'", this.schema.name))
	}
	if err := validateNoDuplicateColumn(this.schema.fields, field); err != nil {
		panic(errors.Wrapf(err, "entity '%s'", this.schema.name))
	}
	if this.schema.fields == nil {
		this.schema.fields = make(map[string]*EntityField)
	}
	this.schema.fields[field.name] = field
	this.schema.fieldsOrder = append(this.schema.fieldsOrder, field.name)

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

func (this *EntitySchemaBuilder) addImplicitEdgeField(rel *EntityRelation) {
	isArray := rel.RelationType == RelationTypeOneToMany || rel.RelationType == RelationTypeManyToMany
	dataType := FieldDataType(FieldDataTypeEntity())
	if isArray {
		dataType = dataType.ArrayType()
	}
	this.Field(DefineField().Name(rel.Edge).DataType(dataType))
}

func (this *EntitySchemaBuilder) TableName(tableName string) *EntitySchemaBuilder {
	this.schema.tableName = tableName
	return this
}

func (this *EntitySchemaBuilder) CompositeUnique(composite ...string) *EntitySchemaBuilder {
	if len(composite) > 0 {
		this.schema.compositeUniques = append(this.schema.compositeUniques, composite)
	}
	return this
}

func (this *EntitySchemaBuilder) SetCompositeUniques(allUniques [][]string) *EntitySchemaBuilder {
	this.schema.compositeUniques = allUniques
	return this
}

func (this *EntitySchemaBuilder) Build() *EntitySchema {
	schema := &this.schema
	if err := populateDbMetadata(schema); err != nil {
		panic(err)
	}
	return schema
}

func validateFieldName(field *EntityField) error {
	if field.name == "" {
		return errors.Errorf("field name is required")
	}
	return nil
}

func validateFieldKeyFlags(field *EntityField) error {
	if field.isPrimaryKey && field.isTenantKey {
		return errors.Errorf("field '%s': isPrimaryKey and isTenantKey are mutually exclusive", field.name)
	}
	return nil
}

func validateSingleTenantKey(existingFields map[string]*EntityField, newField *EntityField) error {
	if !newField.isTenantKey {
		return nil
	}
	for _, f := range existingFields {
		if f != nil && f.isTenantKey {
			return errors.Errorf("field '%s' cannot be tenant key: '%s' is already the tenant key", newField.name, f.name)
		}
	}
	return nil
}

func validateNoDuplicateColumn(existingFields map[string]*EntityField, newField *EntityField) error {
	columnName := newField.name
	for _, f := range existingFields {
		if f != nil && f.name == columnName {
			return errors.Errorf("duplicate column '%s'", columnName)
		}
	}
	return nil
}

type FieldBuilder struct {
	field *EntityField
}

func DefineField() *FieldBuilder {
	return &FieldBuilder{
		field: &EntityField{},
	}
}

func (this *FieldBuilder) Name(name string) *FieldBuilder {
	this.field.name = strings.TrimSpace(name)
	return this
}

func (this *FieldBuilder) Label(label model.LangJson) *FieldBuilder {
	this.field.label = label
	return this
}

func (this *FieldBuilder) IsRequired(isRequired bool) *FieldBuilder {
	this.field.isRequired = isRequired
	return this
}

func (this *FieldBuilder) Required() *FieldBuilder {
	this.field.isRequired = true
	return this
}

func (this *FieldBuilder) PrimaryKey() *FieldBuilder {
	this.field.isPrimaryKey = true
	this.Required()
	return this
}

func (this *FieldBuilder) TenantKey() *FieldBuilder {
	this.field.isTenantKey = true
	return this
}

func (this *FieldBuilder) Unique() *FieldBuilder {
	this.field.isUnique = true
	return this
}

func (this *FieldBuilder) LabelRef(key string) *FieldBuilder {
	return this.Label(model.LangJson{model.LabelRefLanguageCode: key})
}

func (this *FieldBuilder) Description(description model.LangJson) *FieldBuilder {
	this.field.description = description
	return this
}

func (this *FieldBuilder) DataType(dataType FieldDataType) *FieldBuilder {
	this.field.dataType = dataType
	this.field.isArray = dataType.IsArray()
	return this
}

func (this *FieldBuilder) Rule(rule FieldRule) *FieldBuilder {
	rules := this.field.rules
	rules = append(rules, &rule)
	this.field.rules = rules
	return this
}

func (this *FieldBuilder) Default(value any) *FieldBuilder {
	v := value
	this.field.defaultValue = &v
	return this
}

func (this *FieldBuilder) Foreign(relationBuilder *RelationBuilder) *FieldBuilder {
	this.field.relation = relationBuilder.Build()
	return this
}

func (this *FieldBuilder) Build() *EntityField {
	if this.field.name == "" {
		panic("field name is required")
	}
	return this.field
}

type RelationBuilder struct {
	relation *EntityRelation
}

func Edge(edgeName string) *RelationBuilder {
	return &RelationBuilder{
		relation: &EntityRelation{
			Edge: edgeName,
		},
	}
}

func (this *RelationBuilder) Label(label model.LangJson) *RelationBuilder {
	this.relation.label = label
	return this
}

func (this *RelationBuilder) OneToOne(entityName string, destField string) *RelationBuilder {
	this.relation.RelationType = RelationTypeOneToOne
	this.relation.DestEntityName = entityName
	this.relation.DestField = destField
	return this
}

func (this *RelationBuilder) OneToMany(entityName string, destField string) *RelationBuilder {
	this.relation.RelationType = RelationTypeOneToMany
	this.relation.DestEntityName = entityName
	this.relation.DestField = destField
	return this
}

func (this *RelationBuilder) ManyToOne(entityName string, targetField string) *RelationBuilder {
	this.relation.RelationType = RelationTypeManyToOne
	this.relation.DestEntityName = entityName
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

func (this *RelationBuilder) Build() *EntityRelation {
	return this.relation
}
