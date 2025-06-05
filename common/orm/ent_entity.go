package orm

import (
	"reflect"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/util"
)

var registry = &entityRegistry{
	entities: make(map[string]*EntityDescriptor),
}

type entityRegistry struct {
	entities map[string]*EntityDescriptor
}

func RegisterEntity(entity string, descriptor *EntityDescriptor) error {
	if _, ok := registry.entities[entity]; ok {
		return errors.Errorf("entity '%s' already registered", entity)
	}

	registry.entities[entity] = descriptor
	return nil
}

func GetEntity(entity string) (*EntityDescriptor, bool) {
	descriptor, ok := registry.entities[entity]
	return descriptor, ok
}

func DescribeEntity(entity string) *EntityDescriptorBuilder {
	return &EntityDescriptorBuilder{
		descriptor: &EntityDescriptor{
			Entity: entity,
			Edges:  make(map[string]EdgePredicate),
			Fields: make(map[string]reflect.Type),
		},
	}
}

type EntityDescriptor struct {
	Entity string
	Edges  map[string]EdgePredicate
	Fields map[string]reflect.Type
}

func (this *EntityDescriptor) FieldType(field string) (reflect.Type, error) {
	fieldType, ok := this.Fields[field]
	if !ok {
		return nil, errors.Errorf("invalid field '%s' of entity '%s'", field, this.Entity)
	}

	return fieldType, nil
}

func (this *EntityDescriptor) MatchFieldType(field string, value any) (reflect.Type, error) {
	fieldType, err := this.FieldType(field)
	if err != nil {
		return nil, err
	}

	if !util.IsConvertible(value, fieldType) {
		return nil, errors.Errorf("invalid value '%s' for field '%s' of entity '%s'", value, field, this.Entity)
	}

	return fieldType, nil
}

type EntityDescriptorBuilder struct {
	descriptor *EntityDescriptor
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

func (this *EntityDescriptorBuilder) Descriptor() *EntityDescriptor {
	return this.descriptor
}
