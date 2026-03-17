package orm

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
)

const (
	DialectMySql    = "mysql"
	DialectPostgres = "postgres"
)

func GenCreateSql(registry *schema.EntityRegistry, dialect string) ([]string, error) {
	if registry == nil {
		return nil, errors.New("schema registry is required")
	}
	if dialect != DialectPostgres {
		return nil, errors.Errorf("dialect '%s' is not supported", dialect)
	}
	builder := &PgQueryBuilder{}
	var results []string
	err := registry.ForEachOrder(func(schemaName string, s *schema.EntitySchema) error {
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


func isFkOwnerRelationType(relType schema.RelationType) bool {
	return relType == schema.RelationTypeManyToOne || relType == schema.RelationTypeOneToOne
}

func ValidateRelations(registry *schema.EntityRegistry) error {
	if registry == nil {
		return errors.New("schema registry is required")
	}
	err := registry.ForEach(func(schemaName string, s *schema.EntitySchema) error {
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
	registry *schema.EntityRegistry,
	schemaName string,
	relation schema.EntityRelation,
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
	registry *schema.EntityRegistry,
	schemaName string,
	relation schema.EntityRelation,
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
	registry *schema.EntityRegistry,
	sourceSchema *schema.EntitySchema,
	relation schema.EntityRelation,
) (*schema.EntityField, *schema.EntityField, error) {
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
	relation schema.EntityRelation, sourceField *schema.EntityField, foreignField *schema.EntityField,
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

func validateFieldArrayMatchRelationType(relation schema.EntityRelation, sourceField *schema.EntityField) error {
	switch relation.RelationType {
	case schema.RelationTypeOneToMany:
		if !sourceField.IsArray() {
			return errors.Errorf(
				"relation '%s' expects array source field for type '%s'", relation.Edge, relation.RelationType)
		}
	case schema.RelationTypeOneToOne, schema.RelationTypeManyToOne:
		if sourceField.IsArray() {
			return errors.Errorf(
				"relation '%s' expects non-array source field for type '%s'", relation.Edge, relation.RelationType)
		}
	}
	return nil
}
