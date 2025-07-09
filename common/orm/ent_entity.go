package orm

import (
	"reflect"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/util"
)

var registry = &entityRegistry{
	entities: make(map[string]*EntityDescriptor),
}

type entityRegistry struct {
	entities map[string]*EntityDescriptor
}

func RegisterEntity(descriptor *EntityDescriptor) error {
	return registerEntityAliases(descriptor)
}

func registerEntityAliases(descriptor *EntityDescriptor) error {
	for _, alias := range descriptor.EntityAliases {
		if _, ok := registry.entities[alias]; ok {
			return errors.Errorf("entity '%s' already registered", alias)
		}

		registry.entities[alias] = descriptor
	}
	return nil
}

func GetEntity(entity string) (*EntityDescriptor, bool) {
	descriptor, ok := registry.entities[entity]
	return descriptor, ok
}

func DescribeEntity(entity string) *EntityDescriptorBuilder {
	return &EntityDescriptorBuilder{
		descriptor: &EntityDescriptor{
			EntityAliases: []string{entity},
			Edges:         make(map[string]EdgePredicate),
			Fields:        make(map[string]reflect.Type),
			OrderByEdge:   make(map[string]OrderByEdgeFn),
		},
	}
}

func AllEntities() []string {
	entities := make([]string, 0, len(registry.entities))
	for entityName := range registry.entities {
		entities = append(entities, entityName)
	}
	return entities
}

func AllFields(entityName string) (fields []string, edges []string, isOK bool) {
	entity, ok := GetEntity(entityName)
	if !ok {
		return nil, nil, false
	}

	fields = make([]string, 0, len(entity.Fields))
	for fieldName := range entity.Fields {
		fields = append(fields, fieldName)
	}

	edges = make([]string, 0, len(entity.Edges))
	for edgeName := range entity.Edges {
		edges = append(edges, edgeName)
	}

	return fields, edges, true
}

type OrderByEdgeFn func() *sqlgraph.Step

type EntityDescriptor struct {
	EntityAliases []string
	Edges         map[string]EdgePredicate
	Fields        map[string]reflect.Type
	OrderByEdge   map[string]OrderByEdgeFn
}

func (this *EntityDescriptor) Entity() string {
	return this.EntityAliases[0]
}

func (this *EntityDescriptor) Aliases() []string {
	return this.EntityAliases
}

func (this *EntityDescriptor) FieldType(field string) (reflect.Type, error) {
	fieldType, ok := this.Fields[field]
	if !ok {
		return nil, errors.Errorf("invalid field '%s' of entity '%s'", field, this.Entity())
	}

	return fieldType, nil
}

func (this *EntityDescriptor) EdgePredicate(edgeName string) (EdgePredicate, error) {
	hasEdgeWithFn, ok := this.Edges[edgeName]
	if !ok {
		return nil, errors.Errorf("unrecognized relationship '%s' of entity '%s'", edgeName, this.Entity())
	}

	return hasEdgeWithFn, nil
}

func (this *EntityDescriptor) OrderByEdgeStep(edgeName string) (OrderByEdgeFn, error) {
	stepFn, ok := this.OrderByEdge[edgeName]
	if !ok {
		return nil, errors.Errorf("unrecognized sortable relationship '%s' of entity '%s'", edgeName, this.Entity())
	}

	return stepFn, nil
}

func (this *EntityDescriptor) MatchFieldType(field string, value any) (reflect.Type, error) {
	fieldType, err := this.FieldType(field)
	if err != nil {
		return nil, err
	}

	if !util.IsConvertible(value, fieldType) {
		return nil, errors.Errorf("invalid value '%s' for field '%s' of entity '%s'", value, field, this.Entity())
	}

	return fieldType, nil
}

type EntityDescriptorBuilder struct {
	descriptor *EntityDescriptor
}

func (this *EntityDescriptorBuilder) Aliases(aliases ...string) *EntityDescriptorBuilder {
	this.descriptor.EntityAliases = append(this.descriptor.EntityAliases, aliases...)
	return this
}

func (this *EntityDescriptorBuilder) Field(name string, field any) *EntityDescriptorBuilder {
	this.descriptor.Fields[name] = reflect.TypeOf(field)
	return this
}

func (this *EntityDescriptorBuilder) Edge(
	name string,
	predicate EdgePredicate,
) *EntityDescriptorBuilder {
	this.descriptor.Edges[name] = predicate
	return this
}

func (this *EntityDescriptorBuilder) OrderByEdge(
	field string,
	stepFn OrderByEdgeFn,
) *EntityDescriptorBuilder {
	this.descriptor.OrderByEdge[field] = stepFn
	return this
}

func (this *EntityDescriptorBuilder) Descriptor() *EntityDescriptor {
	return this.descriptor
}
