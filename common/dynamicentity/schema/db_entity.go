package schema

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/sky-as-code/nikki-erp/common/array"
	"go.bryk.io/pkg/errors"
)

type Column struct {
	Name     string
	Type     string
	Nullable string
}

var timeType = reflect.TypeOf(time.Time{})

type Operator string

const (
	Equals        Operator = "="
	NotEquals     Operator = "!="
	GreaterThan   Operator = ">"
	GreaterEqual  Operator = ">="
	LessThan      Operator = "<"
	LessEqual     Operator = "<="
	Contains      Operator = "*"
	NotContains   Operator = "!*"
	StartsWith    Operator = "^"
	NotStartsWith Operator = "!^"
	EndsWith      Operator = "$"
	NotEndsWith   Operator = "!$"
	In            Operator = "in"
	NotIn         Operator = "not_in"
	IsSet         Operator = "is_set"
	IsNotSet      Operator = "not_set"
)

type Condition []any

func NewCondition(field string, operator Operator, values ...any) Condition {
	arr := make([]any, 0, 2+len(values))
	arr = append(arr, field, operator)
	arr = append(arr, values...)
	return Condition(arr)
}

func (c Condition) Field() string {
	return fmt.Sprint(c[0])
}

func (c Condition) Operator() Operator {
	if len(c) < 2 {
		return ""
	}
	if op, ok := c[1].(Operator); ok {
		return op
	}
	return Operator(fmt.Sprint(c[1]))
}

func (c Condition) Value() any {
	if len(c) < 3 {
		return nil
	}
	return c[2]
}

func (c Condition) Values() []any {
	if len(c) <= 2 {
		return nil
	}
	return c[2:]
}

type OrderDirection string

const (
	Asc  OrderDirection = "asc"
	Desc OrderDirection = "desc"
)

type SearchOrderItem []string

func (item SearchOrderItem) Direction() OrderDirection {
	if len(item) == 2 {
		return OrderDirection(strings.ToLower(item[1]))
	}
	return Asc
}

type SearchOrder []SearchOrderItem

type SearchGraph struct {
	Condition *Condition   `json:"if,omitempty"`
	And       []SearchNode `json:"and,omitempty"`
	Or        []SearchNode `json:"or,omitempty"`
	Order     SearchOrder  `json:"order,omitempty"`
}

type SearchNode struct {
	Condition *Condition   `json:"if,omitempty"`
	And       []SearchNode `json:"and,omitempty"`
	Or        []SearchNode `json:"or,omitempty"`
}

func NewPgDbEntity(schema *EntitySchema, tableName string) (entity *DbEntity, err error) {
	return NewDbEntity(schema, tableName, resolvePostgresType)
}

func resolvePostgresType(field *EntityField) (string, error) {
	switch field.DataType() {
	case FieldDataTypeEmail,
		FieldDataTypePhone,
		FieldDataTypeString,
		FieldDataTypeSecret,
		FieldDataTypeUrl,
		FieldDataTypeEnumString:
		return "character varying", nil
	case FieldDataTypeUlid:
		return "character varying", nil
	case FieldDataTypeUuid:
		return "uuid", nil
	case FieldDataTypeInteger,
		FieldDataTypeEnumNumber:
		return "integer", nil
	case FieldDataTypeFloat:
		return "double precision", nil
	case FieldDataTypeBoolean:
		return "boolean", nil
	case FieldDataTypeDate:
		return "date", nil
	case FieldDataTypeTime:
		return "time without time zone", nil
	case FieldDataTypeDateTime:
		return "timestamptz", nil
	default:
		return "", fmt.Errorf("unsupported field data type '%s'", field.DataType())
	}
}

func NewDbEntity(schema *EntitySchema, tableName string, resolveDbType ResolveDbType) (entity *DbEntity, err error) {
	name, err := requireName(schema.Name())
	if err != nil {
		return nil, err
	}

	columns, columnSet, primaryKeys, tenantKey, fieldUnique, err := buildColumns(schema.Fields(), name, resolveDbType)
	if err != nil {
		return nil, err
	}
	entityUnique, err := extractEntityUnique(schema.Rules(), columnSet)
	if err != nil {
		return nil, err
	}
	if len(primaryKeys) == 0 {
		return nil, fmt.Errorf("entity '%s' must define at least one primary key column", name)
	}

	entity = &DbEntity{
		name:        name,
		tableName:   tableName,
		columns:     columns,
		primaryKeys: primaryKeys,
		uniqueKeys:  deduplicateKeySets(append(fieldUnique, entityUnique...)),
	}
	if tenantKey != "" {
		entity.tenantKey = &tenantKey
	}
	return entity, nil
}

type DbEntity struct {
	name        string
	tableName   string
	columns     []Column
	primaryKeys []string
	tenantKey   *string
	uniqueKeys  [][]string
}

type ResolveDbType func(field *EntityField) (string, error)

func (this DbEntity) TenantKey() string {
	if this.TenantKey == nil {
		return ""
	}
	return *this.tenantKey
}

func (this DbEntity) Column(name string) (Column, bool) {
	for _, col := range this.columns {
		if col.Name == name {
			return col, true
		}
	}
	return Column{}, false
}

func (this DbEntity) KeyColumns() []string {
	keys := append([]string{}, this.primaryKeys...)
	if tk := this.TenantKey(); tk != "" && !array.Contains(keys, tk) {
		keys = append(keys, tk)
	}
	return keys
}

func (this DbEntity) IsPrimaryKey(name string) bool {
	return array.Contains(this.primaryKeys, name)
}

func (this DbEntity) IsTenantKey(name string) bool {
	return this.TenantKey() == name
}

func (this DbEntity) TableName() string {
	return this.tableName
}

func (this DbEntity) Name() string {
	return this.name
}

func (this DbEntity) Columns() []Column {
	return this.columns
}

func (this DbEntity) PrimaryKeys() []string {
	return this.primaryKeys
}

func (this DbEntity) UniqueKeys() [][]string {
	return this.uniqueKeys
}

func (c Column) IsNullable() bool {
	return !strings.EqualFold(c.Nullable, "NOT NULL")
}

func requireName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", errors.New("entity name is required")
	}
	return name, nil
}

func buildColumns(
	fields map[string]*EntityField,
	entityName string,
	resolveDbType ResolveDbType,
) (columns []Column, columnSet map[string]struct{}, primary []string, tenant string, uniques [][]string, err error) {
	columns = make([]Column, 0, len(fields))
	columnSet = make(map[string]struct{}, len(fields))
	uniques = make([][]string, 0)

	for _, field := range fields {
		if field == nil {
			continue
		}
		if err := validateFieldName(field); err != nil {
			return nil, nil, nil, "", nil, fmt.Errorf("entity '%s': %w", entityName, err)
		}
		columnName := strings.TrimSpace(field.Name())
		if columnName == "" {
			return nil, nil, nil, "", nil, fmt.Errorf("entity '%s' contains a field with empty name", entityName)
		}
		if _, exists := columnSet[columnName]; exists {
			return nil, nil, nil, "", nil, fmt.Errorf("entity '%s' has duplicate column '%s'", entityName, columnName)
		}

		col, isUnique, isPrimary, isTenant, columnErr := makeColumn(field, columnName, entityName, resolveDbType)
		if columnErr != nil {
			return nil, nil, nil, "", nil, columnErr
		}

		columns = append(columns, *col)
		columnSet[columnName] = struct{}{}

		if isUnique {
			uniques = append(uniques, []string{columnName})
		}
		if isPrimary {
			primary = append(primary, columnName)
		}
		if isTenant {
			if tenant != "" && tenant != columnName {
				return nil, nil, nil, "", nil, errors.Errorf("%s must not be a tenant key because %s is already one", columnName, tenant)
			}
			tenant = columnName
		}
	}

	return columns, columnSet, primary, tenant, uniques, nil
}

func makeColumn(
	field *EntityField,
	columnName string,
	entityName string,
	resolveDbType ResolveDbType,
) (column *Column, unique bool, primary bool, tenant bool, err error) {
	columnType, err := resolveDbType(field)
	if err != nil {
		return nil, false, false, false, fmt.Errorf("entity '%s': %w", entityName, err)
	}

	nullable := "NULL"
	if field.IsRequired() {
		nullable = "NOT NULL"
	}

	for _, rule := range field.Rules() {
		ruleName := rule.RuleName()
		if ruleName == "" {
			continue
		}

		switch ruleName {
		case FieldRuleUniqueType:
			unique = true
		case FieldRulePrimaryType:
			primary = true
		case FieldRuleTenantType:
			tenant = true
		}
	}

	column = &Column{Name: columnName, Type: columnType, Nullable: nullable}
	return column, unique, primary, tenant, nil
}

func extractEntityUnique(rules []EntityRule, columnSet map[string]struct{}) (uniques [][]string, err error) {
	uniques = make([][]string, 0)
	for _, rule := range rules {
		if len(rule) == 0 {
			continue
		}
		if fmt.Sprint(rule[0]) != string(EntityRuleNameUnique) {
			continue
		}
		groups, groupErr := collectRuleGroups(rule[1:], columnSet)
		if groupErr != nil {
			return nil, groupErr
		}
		uniques = append(uniques, groups...)
	}
	return uniques, nil
}

func collectRuleGroups(args []any, columnSet map[string]struct{}) ([][]string, error) {
	if len(args) == 0 {
		return nil, nil
	}

	groups := make([][]string, 0, len(args))
	for _, raw := range expandRuleArguments(args) {
		group := make([]string, 0, len(raw))
		for _, item := range raw {
			name := strings.TrimSpace(fmt.Sprint(item))
			if name == "" {
				continue
			}
			if _, ok := columnSet[name]; !ok {
				return nil, fmt.Errorf("unknown column reference '%s'", name)
			}
			group = append(group, name)
		}
		if len(group) == 0 {
			continue
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func expandRuleArguments(args []any) [][]any {
	result := make([][]any, 0, len(args))
	for _, arg := range args {
		switch v := arg.(type) {
		case []any:
			result = append(result, v)
		case []string:
			row := make([]any, len(v))
			for i, item := range v {
				row[i] = item
			}
			result = append(result, row)
		default:
			result = append(result, []any{arg})
		}
	}
	return result
}

func deduplicateKeySets(keys [][]string) [][]string {
	if len(keys) == 0 {
		return keys
	}

	seen := make(map[string]struct{}, len(keys))
	result := make([][]string, 0, len(keys))

	for _, key := range keys {
		if len(key) == 0 {
			continue
		}

		normalized := append([]string(nil), key...)
		sort.Strings(normalized)
		signature := strings.Join(normalized, "|")

		if _, exists := seen[signature]; exists {
			continue
		}

		seen[signature] = struct{}{}
		result = append(result, key)
	}

	return result
}
