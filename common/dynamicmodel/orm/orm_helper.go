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
		return nil, errors.New("GenCreateSql: schema registry is required")
	}
	if dialect != DialectPostgres {
		return nil, errors.Errorf("GenCreateSql: dialect '%s' is not supported", dialect)
	}
	builder := NewPgQueryBuilder()
	var results []string
	err := registry.ForEachOrder(func(schemaName string, s *model.ModelSchema) error {
		sqlParts, clientErrs, genErr := builder.SqlCreateTable(s, registry)
		if genErr != nil {
			return errors.Wrapf(genErr, "GenCreateSql: schema '%s'", schemaName)
		}
		if clientErrs != nil && clientErrs.Count() > 0 {
			return errors.Errorf("GenCreateSql: schema '%s': %v", schemaName, *clientErrs)
		}
		results = append(results, sqlParts...)
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
		return errors.New("ValidateRelations: schema registry is required")
	}
	err := registry.ForEach(func(schemaName string, s *model.ModelSchema) error {
		for _, relation := range s.ToRelations() {
			if err := validateRelation(registry, schemaName, relation); err != nil {
				return errors.Wrapf(err, "ValidateRelations: schema '%s' relation '%s'", schemaName, relation.Edge)
			}
		}
		for _, relation := range s.FromRelations() {
			if relation.RelationType == model.RelationTypeManyToMany {
				return errors.Errorf(
					"ValidateRelations: schema '%s' from-relation '%s': many:many must use EdgeTo, not EdgeFrom",
					schemaName, relation.Edge,
				)
			}
			if err := validateRelation(registry, schemaName, relation); err != nil {
				return errors.Wrapf(err, "ValidateRelations: schema '%s' from-relation '%s'", schemaName, relation.Edge)
			}
			if err := validateFromRelationHasPeerForward(registry, schemaName, relation); err != nil {
				return errors.Wrapf(err, "ValidateRelations: schema '%s' from-relation '%s'", schemaName, relation.Edge)
			}
		}
		return nil
	})
	return err
}

func inverseRelationTypesMatch(forward, inverse model.RelationType) bool {
	switch forward {
	case model.RelationTypeOneToMany:
		return inverse == model.RelationTypeManyToOne
	case model.RelationTypeManyToOne:
		return inverse == model.RelationTypeOneToMany
	case model.RelationTypeOneToOne:
		return inverse == model.RelationTypeOneToOne
	default:
		return false
	}
}

func validateFromRelationHasPeerForward(
	registry *model.SchemaRegistry, inverseSchemaName string, inv model.ModelRelation,
) error {
	peer := registry.Get(inv.DestSchemaName)
	if peer == nil {
		return errors.Errorf(
			"validateFromRelationHasPeerForward: peer schema '%s' not found", inv.DestSchemaName,
		)
	}
	for _, fwd := range peer.ToRelations() {
		if fwd.RelationType == model.RelationTypeManyToMany {
			continue
		}
		if fwd.DestSchemaName != inverseSchemaName {
			continue
		}
		if !inverseRelationTypesMatch(fwd.RelationType, inv.RelationType) {
			continue
		}
		if model.RelationsShareForeignKeyColumns(fwd, inv) {
			return nil
		}
	}
	return errors.Errorf(
		"validateFromRelationHasPeerForward: no matching EdgeTo on peer schema '%s' for the EdgeFrom on '%s'",
		inv.DestSchemaName, inverseSchemaName,
	)
}

func validateRelation(
	registry *model.SchemaRegistry,
	schemaName string,
	relation model.ModelRelation,
) error {
	if relation.RelationType == model.RelationTypeManyToMany {
		return validateManyToManyRelation(registry, schemaName, relation)
	}
	if err := validateRelationInput(registry, schemaName, relation); err != nil {
		return err
	}
	owner := registry.Get(schemaName)
	if err := validateForeignKeyPairsForRelation(registry, owner, relation); err != nil {
		return err
	}
	if err := validateImplicitEdgeFieldForRelation(owner, relation); err != nil {
		return err
	}
	return nil
}

func validateManyToManyRelation(
	registry *model.SchemaRegistry, schemaName string, relation model.ModelRelation,
) error {
	if registry == nil {
		return errors.New("validateManyToManyRelation: schema registry is required")
	}
	sourceSchema := registry.Get(schemaName)
	if sourceSchema == nil {
		return errors.Errorf("validateManyToManyRelation: schema '%s' not found", schemaName)
	}
	if relation.M2mThroughSchemaName == "" || relation.M2mSrcFieldPrefix == "" || relation.DestSchemaName == "" {
		return errors.New(
			"validateManyToManyRelation: ThroughSchemaName, SrcFieldPrefix, DestSchemaName (peer) are required",
		)
	}
	if relation.M2mDestFieldPrefix == "" {
		return errors.New(
			"validateManyToManyRelation: unresolved many-to-many; must call SchemaRegistry.FinalizeRelations",
		)
	}
	edgeField, ok := sourceSchema.Field(relation.Edge)
	if !ok {
		return errors.Errorf(
			"validateManyToManyRelation: edge field '%s' not found on schema '%s'",
			relation.Edge, schemaName,
		)
	}
	if err := validateFieldArrayMatchRelationType(relation, edgeField); err != nil {
		return err
	}
	through := registry.Get(relation.M2mThroughSchemaName)
	if through == nil {
		return errors.Errorf(
			"validateManyToManyRelation: through schema '%s' not found", relation.M2mThroughSchemaName,
		)
	}
	peerSchema := registry.Get(relation.DestSchemaName)
	if peerSchema == nil {
		return errors.Errorf(
			"validateManyToManyRelation: peer schema '%s' not found", relation.DestSchemaName,
		)
	}
	for _, pk := range sourceSchema.PrimaryKeys() {
		col := model.PrefixedThroughColumn(relation.M2mSrcFieldPrefix, pk)
		srcF := sourceSchema.MustField(pk)
		throughF, ok := through.Field(col)
		if !ok {
			return errors.Errorf(
				"validateManyToManyRelation: junction column '%s' not found on '%s'",
				col, relation.M2mThroughSchemaName,
			)
		}
		if err := validateFieldDataTypeMatch(relation, srcF, throughF); err != nil {
			return errors.Wrap(err, "validateManyToManyRelation: src to junction")
		}
	}
	for _, pk := range peerSchema.PrimaryKeys() {
		col := model.PrefixedThroughColumn(relation.M2mDestFieldPrefix, pk)
		peerF := peerSchema.MustField(pk)
		throughF, ok := through.Field(col)
		if !ok {
			return errors.Errorf(
				"validateManyToManyRelation: junction column '%s' not found on '%s'",
				col, relation.M2mThroughSchemaName,
			)
		}
		if err := validateFieldDataTypeMatch(relation, peerF, throughF); err != nil {
			return errors.Wrap(err, "validateManyToManyRelation: peer to junction")
		}
	}
	srcTk := sourceSchema.TenantKey()
	if srcTk != "" {
		tcol := srcTk
		tf, ok := through.Field(tcol)
		if !ok {
			return errors.Errorf(
				"validateManyToManyRelation: junction tenant column '%s' missing on '%s'",
				tcol, relation.M2mThroughSchemaName,
			)
		}
		srcTF := sourceSchema.MustField(srcTk)
		if tf.DataType().String() != srcTF.DataType().String() {
			return errors.Errorf(
				"validateManyToManyRelation: tenant column '%s' type mismatch", tcol,
			)
		}
	}
	return nil
}

func validateRelationInput(
	registry *model.SchemaRegistry,
	schemaName string,
	relation model.ModelRelation,
) error {
	if registry == nil {
		return errors.New("validateRelationInput: schema registry is required")
	}
	if schemaName == "" {
		return errors.New("validateRelationInput: schema name is required")
	}
	if registry.Get(schemaName) == nil {
		return errors.Errorf("validateRelationInput: schema '%s' not found in registry", schemaName)
	}
	if relation.DestSchemaName == "" {
		return errors.New("validateRelationInput: relation destination schema name is required")
	}
	if relation.InversePeerSchemaName != "" {
		return errors.New(
			"validateRelationInput: unresolved EdgeFrom peer relation; call SchemaRegistry.FinalizeRelations",
		)
	}
	if len(relation.EffectiveForeignKeys()) == 0 {
		return errors.New("validateRelationInput: foreign key column mapping is required")
	}
	return nil
}

func validateForeignKeyPairsForRelation(
	registry *model.SchemaRegistry, owner *model.ModelSchema, relation model.ModelRelation,
) error {
	if owner == nil {
		return errors.New("validateForeignKeyPairsForRelation: owner schema is required")
	}
	dest := registry.Get(relation.DestSchemaName)
	if dest == nil {
		return errors.Errorf(
			"validateForeignKeyPairsForRelation: referenced schema '%s' not found",
			relation.DestSchemaName,
		)
	}
	for _, pair := range relation.EffectiveForeignKeys() {
		if err := validateSingleForeignKeyPair(owner, dest, relation, pair); err != nil {
			return err
		}
	}
	return nil
}

func validateSingleForeignKeyPair(
	owner, dest *model.ModelSchema, relation model.ModelRelation, pair model.ForeignKeyColumnPair,
) error {
	fkSchema, refSchema := fkAndReferencedSchemas(owner, dest, relation.RelationType)
	fkField, ok := fkSchema.Field(pair.FkColumn)
	if !ok {
		return errors.Errorf(
			"validateSingleForeignKeyPair: FK column '%s' not found on schema '%s'",
			pair.FkColumn, fkSchema.Name(),
		)
	}
	refField, ok := refSchema.Field(pair.ReferencedColumn)
	if !ok {
		return errors.Errorf(
			"validateSingleForeignKeyPair: referenced column '%s' not found on schema '%s'",
			pair.ReferencedColumn, refSchema.Name(),
		)
	}
	if fkField.DataType().String() != refField.DataType().String() {
		return errors.Errorf(
			"validateSingleForeignKeyPair: relation '%s': FK column '%s' type '%s' vs referenced '%s' type '%s'",
			relation.Edge, pair.FkColumn, fkField.DataType().String(),
			pair.ReferencedColumn, refField.DataType().String(),
		)
	}
	if !refSchema.IsPrimaryKey(pair.ReferencedColumn) {
		return errors.Errorf(
			"validateSingleForeignKeyPair: '%s' is not a primary key on schema '%s'",
			pair.ReferencedColumn, refSchema.Name(),
		)
	}
	return nil
}

func fkAndReferencedSchemas(
	owner, dest *model.ModelSchema, relType model.RelationType,
) (fkSchema *model.ModelSchema, refSchema *model.ModelSchema) {
	switch relType {
	case model.RelationTypeManyToOne, model.RelationTypeOneToOne:
		return owner, dest
	case model.RelationTypeOneToMany:
		return dest, owner
	default:
		return owner, dest
	}
}

func validateImplicitEdgeFieldForRelation(owner *model.ModelSchema, relation model.ModelRelation) error {
	if relation.Edge == "" {
		return errors.New("validateImplicitEdgeFieldForRelation: relation edge name is required")
	}
	field, ok := owner.Field(relation.Edge)
	if !ok {
		return errors.Errorf(
			"validateImplicitEdgeFieldForRelation: edge field '%s' not found on schema '%s'",
			relation.Edge, owner.Name(),
		)
	}
	return validateFieldArrayMatchRelationType(relation, field)
}

func validateFieldDataTypeMatch(
	relation model.ModelRelation, leftField *model.ModelField, rightField *model.ModelField,
) error {
	if leftField.DataType().String() != rightField.DataType().String() {
		return errors.Errorf(
			"validateFieldDataTypeMatch: relation '%s': field '%s' type '%s' vs field '%s' type '%s'",
			relation.Edge, leftField.Name(), leftField.DataType().String(),
			rightField.Name(), rightField.DataType().String(),
		)
	}
	return nil
}

func validateFieldArrayMatchRelationType(relation model.ModelRelation, sourceField *model.ModelField) error {
	switch relation.RelationType {
	case model.RelationTypeManyToMany:
		if !sourceField.IsArray() {
			return errors.Errorf(
				"validateFieldArrayMatchRelationType: relation '%s' expects array model field for many:many",
				relation.Edge,
			)
		}
	case model.RelationTypeOneToMany:
		if !sourceField.IsArray() {
			return errors.Errorf(
				"validateFieldArrayMatchRelationType: relation '%s' expects array source field for type '%s'",
				relation.Edge, relation.RelationType)
		}
	case model.RelationTypeOneToOne, model.RelationTypeManyToOne:
		if sourceField.IsArray() {
			return errors.Errorf(
				"validateFieldArrayMatchRelationType: relation '%s' expects non-array source field for type '%s'",
				relation.Edge, relation.RelationType)
		}
	}
	return nil
}
