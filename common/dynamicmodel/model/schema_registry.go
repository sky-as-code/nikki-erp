package model

import (
	"sort"
	"sync"

	"go.bryk.io/pkg/errors"
)

var schemaRegistry = &SchemaRegistry{
	schemas: make(map[string]*ModelSchema),
	mu:      &sync.RWMutex{},
}

type SchemaRegistry struct {
	schemas      map[string]*ModelSchema
	orderedNames []string
	mu           *sync.RWMutex
}

type RelationValidator func(registry *SchemaRegistry, schemaName string, relation ModelRelation) error

func (this *SchemaRegistry) Get(name string) *ModelSchema {
	this.mu.RLock()
	defer this.mu.RUnlock()
	schema, exists := this.schemas[name]
	if !exists {
		return nil
	}
	return schema
}

func (this *SchemaRegistry) FieldSafe(schemaName string, fieldName string) (*ModelField, error) {
	schema := this.Get(schemaName)
	if schema == nil {
		return nil, errors.Errorf("schema '%s' not found", schemaName)
	}

	field, ok := schema.Field(fieldName)
	if !ok {
		return nil, errors.Errorf("field '%s' not found in schema '%s'", fieldName, schemaName)
	}

	return field, nil
}

func (this *SchemaRegistry) Field(schemaName string, fieldName string) *ModelField {
	field, err := this.FieldSafe(schemaName, fieldName)
	if err != nil {
		panic(err)
	}
	return field
}

// RegisterSchemaB executes the schemaBuilder then registers a schema using its name (set via EntitySchemaBuilder.Name) as the registry key.
// Returns an error if a schema with the same name is already registered.
func RegisterSchemaB(schemaBuilder *EntitySchemaBuilder) error {
	if schemaBuilder == nil {
		return errors.New("schemaBuilder cannot be nil")
	}

	return RegisterSchema(schemaBuilder.Build())
}

// RegisterSchema registers a schema using its name as the registry key.
// Returns an error if a schema with the same name is already registered.
func RegisterSchema(schema *ModelSchema) error {
	name := schema.Name()
	if name == "" {
		return errors.New("schema name must not be empty")
	}

	schemaRegistry.mu.Lock()
	defer schemaRegistry.mu.Unlock()

	if _, exists := schemaRegistry.schemas[name]; exists {
		return errors.Errorf("schema '%s' already registered", name)
	}

	schemaRegistry.schemas[name] = schema
	schemaRegistry.orderedNames = computeTopoOrder(schemaRegistry.schemas)
	return nil
}

func (this *SchemaRegistry) ForEach(fn func(schemaName string, schema *ModelSchema) error) error {
	this.mu.RLock()
	defer this.mu.RUnlock()
	for schemaName, schemaItem := range this.schemas {
		if err := fn(schemaName, schemaItem); err != nil {
			return err
		}
	}
	return nil
}

// ForEachOrder iterates schemas in FK-dependency order (parents before children),
// suitable for generating CREATE TABLE statements in the correct sequence.
func (this *SchemaRegistry) ForEachOrder(fn func(schemaName string, schema *ModelSchema) error) error {
	this.mu.RLock()
	defer this.mu.RUnlock()
	for _, name := range this.orderedNames {
		if err := fn(name, this.schemas[name]); err != nil {
			return err
		}
	}
	return nil
}

// GetSchema retrieves a registered schema by its name.
func GetSchema(name string) *ModelSchema {
	return schemaRegistry.Get(name)
}

// MustGetSchema retrieves a registered schema by its name.
func MustGetSchema(name string) *ModelSchema {
	schema := schemaRegistry.Get(name)
	if schema == nil {
		panic(errors.Errorf("schema '%s' not found", name))
	}
	return schema
}

// GetOrRegisterSchema first attempts to retrieve a registered schema by its name.
// If not found, it builds a new schema using the builder and registers it.
func GetOrRegisterSchema(newSchema *ModelSchema) *ModelSchema {
	name := newSchema.Name()
	schema := schemaRegistry.Get(name)
	if schema == nil {
		RegisterSchema(newSchema)
		schema = schemaRegistry.Get(name)
	}
	return schema
}

func GetSchemaRegistry() *SchemaRegistry {
	return schemaRegistry
}

func CloneField(schema *ModelSchema, fieldName string) *FieldBuilder {
	field := schema.MustField(fieldName)
	clonedField := field.Clone()
	return &FieldBuilder{
		field: clonedField,
	}
}

func CloneFieldN(schemaName string, fieldName string) *FieldBuilder {
	field := schemaRegistry.Field(schemaName, fieldName)
	clonedField := field.Clone()
	return &FieldBuilder{
		field: clonedField,
	}
}

// computeTopoOrder returns canonical schema names sorted by FK dependency order
// (referenced/parent schemas appear before schemas that depend on them).
// Schemas with no FK relationships are sorted alphabetically for determinism.
// Cyclic dependencies are handled gracefully by appending remaining nodes last.
func computeTopoOrder(schemas map[string]*ModelSchema) []string {
	known := buildKnownSet(schemas)
	inDegree, dependents := buildDepGraph(schemas, known)
	return kahnSort(known, inDegree, dependents)
}

// buildKnownSet returns a set of all canonical names currently in the registry,
// used for O(1) membership checks during dependency graph construction.
func buildKnownSet(schemas map[string]*ModelSchema) map[string]bool {
	known := make(map[string]bool, len(schemas))
	for name := range schemas {
		known[name] = true
	}
	return known
}

// buildDepGraph constructs the two data structures required by Kahn's algorithm:
// inDegree counts how many registry schemas each schema depends on, and
// dependents maps each schema to the list of schemas that directly depend on it.
func buildDepGraph(
	schemas map[string]*ModelSchema,
	known map[string]bool,
) (inDegree map[string]int, dependents map[string][]string) {
	inDegree = make(map[string]int, len(schemas))
	dependents = make(map[string][]string, len(schemas))
	for name := range schemas {
		inDegree[name] = 0
	}
	for name, s := range schemas {
		for _, dep := range fkDependencies(s, known) {
			if dep == name {
				continue
			}
			inDegree[name]++
			dependents[dep] = append(dependents[dep], name)
		}
	}
	return inDegree, dependents
}

// fkDependencies returns canonical names of schemas that must be created before s,
// based on FK-owning relations (many:one, one:one) whose destination is in the registry.
func fkDependencies(s *ModelSchema, known map[string]bool) []string {
	var deps []string
	for _, rel := range s.relations {
		if isFkOwnerRelation(rel.RelationType) && known[rel.DestEntityName] {
			deps = append(deps, rel.DestEntityName)
		}
	}
	return deps
}

// isFkOwnerRelation reports whether the given relation type places the FK column
// on the current entity's table (many:one and one:one), as opposed to one:many
// (FK lives on the other table) or many:many (FK lives in a junction table).
func isFkOwnerRelation(relType RelationType) bool {
	return relType == RelationTypeManyToOne || relType == RelationTypeOneToOne
}

// kahnSort runs Kahn's BFS topological sort and returns schemas in dependency
// order. Nodes with equal in-degree are emitted alphabetically for determinism.
func kahnSort(known map[string]bool, inDegree map[string]int, dependents map[string][]string) []string {
	queue := initialQueue(inDegree)
	result := make([]string, 0, len(known))
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)
		nextBatch := drainDependents(node, inDegree, dependents)
		queue = append(queue, nextBatch...)
	}
	return appendCyclicRemainder(result, known)
}

// initialQueue collects all schemas with no dependencies (in-degree zero) and
// sorts them alphabetically to seed the BFS with a deterministic start order.
func initialQueue(inDegree map[string]int) []string {
	queue := make([]string, 0, len(inDegree))
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}
	sort.Strings(queue)
	return queue
}

// drainDependents decrements the in-degree of every schema that depends on node
// and returns those whose in-degree reaches zero, sorted for determinism.
func drainDependents(node string, inDegree map[string]int, dependents map[string][]string) []string {
	batch := make([]string, 0, len(dependents[node]))
	for _, dep := range dependents[node] {
		inDegree[dep]--
		if inDegree[dep] == 0 {
			batch = append(batch, dep)
		}
	}
	sort.Strings(batch)
	return batch
}

// appendCyclicRemainder appends any schemas not yet in result (caused by a
// dependency cycle) in alphabetical order, preventing silent data loss.
func appendCyclicRemainder(result []string, known map[string]bool) []string {
	inResult := make(map[string]bool, len(result))
	for _, name := range result {
		inResult[name] = true
	}
	remainder := make([]string, 0)
	for name := range known {
		if !inResult[name] {
			remainder = append(remainder, name)
		}
	}
	sort.Strings(remainder)
	return append(result, remainder...)
}
