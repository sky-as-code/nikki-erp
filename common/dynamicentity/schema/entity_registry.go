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

func (r *entityRegistry) Add(name string, schema *EntitySchema) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.schemas[name]; exists {
		return fmt.Errorf("schema '%s' already registered", name)
	}

	r.schemas[name] = schema
	return nil
}

func (r *entityRegistry) Get(name string) (*EntitySchema, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schema, ok := r.schemas[name]
	return schema, ok
}

func (r *entityRegistry) FieldSafe(schemaName string, fieldName string) (*EntityField, error) {
	schema, ok := r.Get(schemaName)
	if !ok {
		return nil, fmt.Errorf("schema '%s' not found", schemaName)
	}

	field, ok := schema.Field(fieldName)
	if !ok {
		return nil, fmt.Errorf("field '%s' not found in schema '%s'", fieldName, schemaName)
	}

	return field, nil
}

func (r *entityRegistry) Field(schemaName string, fieldName string) *EntityField {
	field, err := r.FieldSafe(schemaName, fieldName)
	if err != nil {
		panic(err)
	}
	return field
}

func AddSchema(schema *EntitySchema) error {
	return schemaRegistry.Add(schema.Name(), schema)
}

func GetSchema(name string) (*EntitySchema, bool) {
	return schemaRegistry.Get(name)
}

func CloneField(schemaName string, fieldName string) *FieldBuilder {
	field := schemaRegistry.Field(schemaName, fieldName)
	clonedField := field.Clone()
	return &FieldBuilder{
		field: clonedField,
	}
}
