package orm

import (
	"reflect"

	"github.com/sky-as-code/nikki-erp/common/util"
	"go.bryk.io/pkg/errors"
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

func (this *EntityDescriptor) MatchFieldType(field string, value any) (reflect.Type, error) {
	fieldType, ok := this.Fields[field]
	if !ok {
		return nil, errors.Errorf("invalid field '%s' of entity '%s'", field, this.Entity)
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
	var fieldType reflect.Type

	// Get the type of the field
	fieldValue := reflect.ValueOf(field)

	// Check if field is a pointer
	if fieldValue.Kind() == reflect.Ptr {
		// Get the type the pointer points to
		fieldType = fieldValue.Elem().Type()
	} else {
		// Get the type directly
		fieldType = fieldValue.Type()
	}

	this.descriptor.Fields[name] = fieldType
	return this
}

// func (this *EntityDescriptorBuilder) Edge(name string, predicate EdgePredicate) *EntityDescriptorBuilder {
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
