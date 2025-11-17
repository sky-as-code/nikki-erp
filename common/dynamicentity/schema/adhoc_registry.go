package schema

import (
	"fmt"
	"sync"
)

var adhocRegistry = &adhocSchemaRegistry{
	schemas: make(map[string]*AdhocSchema),
	// entitySchemas: make(map[string]*EntitySchema),
	mu: &sync.RWMutex{},
}

func AdhocRegistry() *adhocSchemaRegistry {
	return adhocRegistry
}

type adhocSchemaRegistry struct {
	schemas map[string]*AdhocSchema
	// entitySchemas map[string]*EntitySchema
	mu *sync.RWMutex
}

func (r *adhocSchemaRegistry) Add(name string, schema *AdhocSchema) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.schemas[name]; exists {
		return fmt.Errorf("adhoc schema '%s' already registered", name)
	}

	r.schemas[name] = schema

	// Build entity schema from adhoc schema for validator
	// entitySchema := r.buildEntitySchemaFromAdhoc(schema)
	// r.entitySchemas[name] = entitySchema

	return nil
}

func GetAdhocSchema(name string) (*AdhocSchema, error) {
	return adhocRegistry.GetSchema(name)
}

func (r *adhocSchemaRegistry) GetSchema(name string) (*AdhocSchema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schema, ok := r.schemas[name]
	if !ok {
		return nil, fmt.Errorf("adhoc schema '%s' not found", name)
	}

	return schema, nil
}

// func (r *adhocSchemaRegistry) buildEntitySchemaFromAdhoc(adhoc *AdhocSchema) *EntitySchema {
// 	fields := make(map[string]*EntityField)

// 	for name, adhocField := range adhoc.fields {
// 		if adhocField.isHolder {
// 			// Skip holders - they don't generate validation rules
// 			continue
// 		}

// 		field := adhocField.field
// 		if field == nil {
// 			continue
// 		}
// 		fieldCopy := field.Clone()
// 		if fieldCopy.name == "" {
// 			fieldCopy.name = name
// 		}
// 		if err := validateFieldName(fieldCopy); err != nil {
// 			panic(fmt.Errorf("adhoc schema field '%s': %w", name, err))
// 		}
// 		fields[name] = fieldCopy
// 	}

// 	return &EntitySchema{
// 		name:   "",
// 		fields: fields,
// 	}
// }
