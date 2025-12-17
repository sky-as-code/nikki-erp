package dynamicentity

import (
	"sync"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicentity/model"
	// dorm "github.com/sky-as-code/nikki-erp/common/dynamicentity/orm"
	dschema "github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"
)

type ValidatorCreator interface {
	CreateValidator(schema dschema.EntitySchema) Validator
}

type Validator interface {
	ValidateS(entity any) ft.ValidationErrors
	ValidateM(entity dmodel.EntityMap) ft.ValidationErrors
}

type DbEntityCreator interface {
	CreateDbEntity(schema dschema.EntitySchema) dschema.DbEntity
}

type SchemaProvider interface {
	OnSchemaChanged(func(newSchema dschema.EntitySchema))
	GetSchemas(entityNames []string) ([]dschema.EntitySchema, error)
}

type EntityToolbox struct {
	validator Validator
	dbEntity  dschema.DbEntity
	schema    dschema.EntitySchema
}

type NewDynamicEntityEngineOtps struct {
	dig.In

	DbEntityCreator  DbEntityCreator
	SchemaProvider   SchemaProvider
	ValidatorCreator ValidatorCreator
}

func NewDynamicEntityEngine(opts NewDynamicEntityEngineOtps) *DynamicEntityEngine {
	return &DynamicEntityEngine{
		dbEntityCreator:  opts.DbEntityCreator,
		schemaProvider:   opts.SchemaProvider,
		validatorCreator: opts.ValidatorCreator,
		mu:               &sync.RWMutex{},
	}
}

type DynamicEntityEngine struct {
	entities         map[string]EntityToolbox
	dbEntityCreator  DbEntityCreator
	schemaProvider   SchemaProvider
	validatorCreator ValidatorCreator

	mu *sync.RWMutex
}

func (this *DynamicEntityEngine) Prepare(entityNames []string) error {
	if this.entities != nil {
		return nil
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	this.entities = make(map[string]EntityToolbox)
	schemas, err := this.schemaProvider.GetSchemas(entityNames)
	if err != nil {
		return errors.Wrap(err, "failed to get schemas")
	}
	for _, schema := range schemas {
		this.prepareEntity(schema)
	}

	this.schemaProvider.OnSchemaChanged(func(newSchema dschema.EntitySchema) {
		this.prepareEntity(newSchema)
	})

	return nil
}

func (this *DynamicEntityEngine) prepareEntity(schema dschema.EntitySchema) {
	dbEntity := this.dbEntityCreator.CreateDbEntity(schema)
	validator := this.validatorCreator.CreateValidator(schema)
	this.entities[schema.Name()] = EntityToolbox{
		dbEntity:  dbEntity,
		schema:    schema,
		validator: validator,
	}
}

func (this *DynamicEntityEngine) DbEntity(entityName string) (dschema.DbEntity, bool) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	entityToolbox, ok := this.entities[entityName]
	if !ok {
		return dschema.DbEntity{}, false
	}
	return entityToolbox.dbEntity, true
}

func (this *DynamicEntityEngine) Schema(entityName string) (dschema.EntitySchema, bool) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	entityToolbox, ok := this.entities[entityName]
	if !ok {
		return dschema.EntitySchema{}, false
	}
	return entityToolbox.schema, true
}

func (this *DynamicEntityEngine) Validator(entityName string) (Validator, bool) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	entityToolbox, ok := this.entities[entityName]
	if !ok {
		return nil, false
	}
	return entityToolbox.validator, true
}
