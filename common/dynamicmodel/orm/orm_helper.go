package orm

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

const (
	DialectMySql    = "mysql"
	DialectPostgres = "postgres"
)

func GenCreateSql(registry *model.SchemaRegistry, dialect string) ([]string, error) {
	if registry == nil {
		return nil, errors.New("schema registry is required")
	}
	if dialect != DialectPostgres {
		return nil, errors.Errorf("dialect '%s' is not supported", dialect)
	}
	builder := &PgQueryBuilder{}
	var results []string
	err := registry.ForEachOrder(func(schemaName string, s *model.ModelSchema) error {
		sql, genErr := builder.SqlCreateTable(s, registry)
		if genErr != nil {
			return errors.Wrapf(genErr, "schema '%s'", schemaName)
		}
		results = append(results, sql)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func isFkOwnerRelationType(relType model.RelationType) bool {
	return relType == model.RelationTypeManyToOne || relType == model.RelationTypeOneToOne
}

func ValidateRelations(registry *model.SchemaRegistry) error {
	if registry == nil {
		return errors.New("schema registry is required")
	}
	err := registry.ForEach(func(schemaName string, s *model.ModelSchema) error {
		for _, relation := range s.Relations() {
			if err := validateRelation(registry, schemaName, relation); err != nil {
				return errors.Wrapf(err, "schema '%s' relation '%s'", schemaName, relation.Edge)
			}
		}
		return nil
	})
	return err
}

func validateRelation(
	registry *model.SchemaRegistry,
	schemaName string,
	relation model.ModelRelation,
) error {
	if err := validateRelationInput(registry, schemaName, relation); err != nil {
		return err
	}
	sourceSchema := registry.Get(schemaName)
	sourceField, foreignField, err := resolveRelationFields(registry, sourceSchema, relation)
	if err != nil {
		return err
	}
	if err := validateFieldDataTypeMatch(relation, sourceField, foreignField); err != nil {
		return err
	}
	if err := validateFieldArrayMatchRelationType(relation, sourceField); err != nil {
		return err
	}
	return nil
}

func validateRelationInput(
	registry *model.SchemaRegistry,
	schemaName string,
	relation model.ModelRelation,
) error {
	if registry == nil {
		return errors.New("schema registry is required")
	}
	if schemaName == "" {
		return errors.New("schema name is required")
	}
	if registry.Get(schemaName) == nil {
		return errors.Errorf("schema '%s' not found in registry", schemaName)
	}
	if relation.DestEntityName == "" {
		return errors.New("relation destination schema name is required")
	}
	if relation.SrcField == "" || relation.DestField == "" {
		return errors.New("relation source field and destination field are required")
	}
	return nil
}

func resolveRelationFields(
	registry *model.SchemaRegistry,
	sourceSchema *model.ModelSchema,
	relation model.ModelRelation,
) (*model.ModelField, *model.ModelField, error) {
	sourceField, ok := sourceSchema.Field(relation.SrcField)
	if !ok {
		return nil, nil, errors.Errorf(
			"source field '%s' does not exist in schema '%s'", relation.SrcField, sourceSchema.Name())
	}
	foreignSchema := registry.Get(relation.DestEntityName)
	if foreignSchema == nil {
		return nil, nil, errors.Errorf(
			"referenced schema '%s' not found in registry", relation.DestEntityName)
	}
	foreignField, ok := foreignSchema.Field(relation.DestField)
	if !ok {
		return nil, nil, errors.Errorf(
			"referenced field '%s' does not exist in schema '%s'", relation.DestField, relation.DestEntityName)
	}
	return sourceField, foreignField, nil
}

func validateFieldDataTypeMatch(
	relation model.ModelRelation, sourceField *model.ModelField, foreignField *model.ModelField,
) error {
	sourceType := sourceField.DataType().String()
	foreignType := foreignField.DataType().String()
	if sourceType != foreignType {
		return errors.Errorf(
			"relation '%s': source field '%s' has type '%s' but destination field '%s' has type '%s'",
			relation.Edge, relation.SrcField, sourceType, relation.DestField, foreignType)
	}
	return nil
}

func validateFieldArrayMatchRelationType(relation model.ModelRelation, sourceField *model.ModelField) error {
	switch relation.RelationType {
	case model.RelationTypeOneToMany:
		if !sourceField.IsArray() {
			return errors.Errorf(
				"relation '%s' expects array source field for type '%s'", relation.Edge, relation.RelationType)
		}
	case model.RelationTypeOneToOne, model.RelationTypeManyToOne:
		if sourceField.IsArray() {
			return errors.Errorf(
				"relation '%s' expects non-array source field for type '%s'", relation.Edge, relation.RelationType)
		}
	}
	return nil
}
