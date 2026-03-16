package schema

import (
	"fmt"
	"sync"
)

var schemaRegistry = &entityRegistry{
	schemas: make(map[string]*EntitySchema),
	mu:      &sync.RWMutex{},
}

type entityRegistry struct {
	schemas map[string]*EntitySchema
	mu      *sync.RWMutex
}

// CanonicalSchemaName constructs a registry key from module name, optional sub-module name, and schema name.
// Format: "{moduleName}.{subModName if specify}.{schemaName}"
func CanonicalSchemaName(schemaName string, moduleName string, subModName ...string) string {
	if len(subModName) > 0 && subModName[0] != "" {
		return fmt.Sprintf("%s.%s.%s", moduleName, subModName, schemaName)
	}
	return fmt.Sprintf("%s.%s", moduleName, schemaName)
}

// func (this *entityRegistry) Add(name string, schema *EntitySchema) error {
// 	this.mu.Lock()
// 	defer this.mu.Unlock()

// 	if _, exists := this.schemas[name]; exists {
// 		return fmt.Errorf("schema '%s' already registered", name)
// 	}

// 	this.schemas[name] = schema
// 	return nil
// }

func (this *entityRegistry) Get(name string) *EntitySchema {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.schemas[name]
}

func (this *entityRegistry) FieldSafe(schemaName string, fieldName string) (*EntityField, error) {
	schema := this.Get(schemaName)
	if schema == nil {
		return nil, fmt.Errorf("schema '%s' not found", schemaName)
	}

	field, ok := schema.Field(fieldName)
	if !ok {
		return nil, fmt.Errorf("field '%s' not found in schema '%s'", fieldName, schemaName)
	}

	return field, nil
}

func (this *entityRegistry) Field(schemaName string, fieldName string) *EntityField {
	field, err := this.FieldSafe(schemaName, fieldName)
	if err != nil {
		panic(err)
	}
	return field
}

// RegisterSchema registers a schema with the given module name and optional sub-module name.
// It validates relations before registration.
// Registry key format: "{moduleName}.{subModName if specify}.{schemaName}"
func RegisterSchema(schemaBuilder *EntitySchemaBuilder, moduleName string, subModName ...string) error {
	if schemaBuilder == nil {
		return fmt.Errorf("schemaBuilder cannot be nil")
	}

	var subMod string
	if len(subModName) > 0 && subModName[0] != "" {
		subMod = subModName[0]
	}
	schema := schemaBuilder.Build()
	key := CanonicalSchemaName(schema.Name(), moduleName, subMod)

	schemaRegistry.mu.Lock()
	defer schemaRegistry.mu.Unlock()

	if _, exists := schemaRegistry.schemas[key]; exists {
		return fmt.Errorf("schema '%s' already registered", key)
	}

	// Validate relations (pass registry map to avoid deadlock)
	if err := validateRelations(schema, schema, schemaRegistry.schemas); err != nil {
		return fmt.Errorf("schema '%s' has invalid relations: %w", key, err)
	}

	schemaRegistry.schemas[key] = schema
	return nil
}

// validateRelations validates all relations in a schema
func validateRelations(schema *EntitySchema, sourceSchema *EntitySchema, registry map[string]*EntitySchema) error {
	for _, relation := range schema.Relations() {
		if err := validateRelation(relation, sourceSchema, registry); err != nil {
			return err
		}
	}
	return nil
}

// validateRelation validates a single relation
func validateRelation(relation EntityRelation, sourceSchema *EntitySchema, registry map[string]*EntitySchema) error {
	// Check if ForeignEntityName is provided
	if relation.DestEntityName == "" {
		return fmt.Errorf("relation from field '%s' has empty ForeignEntityName", relation.SrcField)
	}

	// Check if ForeignEntityName exists in registry
	foreignSchema := registry[relation.DestEntityName]
	if foreignSchema == nil {
		return fmt.Errorf("relation from field '%s' references non-existent entity '%s'", relation.SrcField, relation.DestEntityName)
	}

	// Check if ForeignField exists in the foreign entity
	foreignField, ok := foreignSchema.Field(relation.DestField)
	if !ok {
		return fmt.Errorf("relation from field '%s' references non-existent field '%s' in entity '%s'", relation.SrcField, relation.DestField, relation.DestEntityName)
	}

	// Get the source field to check data type
	sourceField, ok := sourceSchema.Field(relation.SrcField)
	if !ok {
		return fmt.Errorf("source field '%s' not found in schema '%s'", relation.SrcField, sourceSchema.Name())
	}

	// Check if FieldDataType matches (by type name, since factory creates new instances)
	if sourceField.DataType().String() != foreignField.DataType().String() {
		return fmt.Errorf("relation from field '%s' (type: %s) to field '%s' (type: %s) in entity '%s': data types do not match",
			relation.SrcField, sourceField.DataType().String(), relation.DestField, foreignField.DataType().String(), relation.DestEntityName)
	}

	return nil
}

// AddSchema registers a schema with the given module name and optional sub-module name.
// Registry key format: "{moduleName}.{subModName if specify}.{schemaName}"
// func addSchema(schema *EntitySchema, moduleName string, subModName ...string) error {
// 	var subMod string
// 	if len(subModName) > 0 && subModName[0] != "" {
// 		subMod = subModName[0]
// 	}
// 	key := buildCanonicalName(moduleName, subMod, schema.Name())
// 	return schemaRegistry.Add(key, schema)
// }

// GetSchema retrieves a schema by canonical schema name.
// Canonical name format: "{moduleName}.{subModName if specify}.{schemaName}"
func GetSchema(canonicalName string) *EntitySchema {
	return schemaRegistry.Get(canonicalName)
}

// GetSchemaM retrieves a schema by schema name, module name, optional sub-module name.
func GetSchemaM(schemaName string, moduleName string, subModName ...string) *EntitySchema {
	var subMod string
	if len(subModName) > 0 && subModName[0] != "" {
		subMod = subModName[0]
	}
	key := CanonicalSchemaName(schemaName, moduleName, subMod)
	return schemaRegistry.Get(key)
}

func CloneField(schemaName string, fieldName string) *FieldBuilder {
	field := schemaRegistry.Field(schemaName, fieldName)
	clonedField := field.Clone()
	return &FieldBuilder{
		field: clonedField,
	}
}

// CloneFieldWithModule clones a field from a schema identified by module and optional sub-module.
func CloneFieldWithModule(schemaName string, moduleName string, fieldName string, subModName ...string) *FieldBuilder {
	var subMod string
	if len(subModName) > 0 && subModName[0] != "" {
		subMod = subModName[0]
	}
	key := CanonicalSchemaName(schemaName, moduleName, subMod)
	field := schemaRegistry.Field(key, fieldName)
	clonedField := field.Clone()
	return &FieldBuilder{
		field: clonedField,
	}
}
