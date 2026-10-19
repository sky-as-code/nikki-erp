// Package defaults reads a module's org defaults: one `<schema_name>.json` file per resource,
// holding an array of rows that the module writes through its domain services.
//
// A row cannot know the ids of the rows it references, because the domain service generates
// them. Instead a row names itself with "$alias" and other rows reference it with "@alias":
//
//	[{"$alias": "wh_main", "code": "WH", ...}]
//	[{"code": "Stock", "warehouse_id": "@wh_main", ...}]
//
// The module decides the write order; Aliases only remembers which id each alias received.
package defaults

import (
	"encoding/json"
	"io/fs"
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	AliasKey  = "$alias"
	RefPrefix = "@"
)

// Load reads `<schemaName>.json` from fsys.
func Load(fsys fs.FS, schemaName string) ([]dmodel.DynamicFields, error) {
	content, err := fs.ReadFile(fsys, schemaName+".json")
	if err != nil {
		return nil, errors.Wrapf(err, "defaults: reading %s.json", schemaName)
	}
	rows := []dmodel.DynamicFields{}
	if err := json.Unmarshal(content, &rows); err != nil {
		return nil, errors.Wrapf(err, "defaults: parsing %s.json", schemaName)
	}
	return rows, nil
}

// Aliases maps each alias to the id its row was created with.
type Aliases struct {
	ids map[string]model.Id
}

func NewAliases() *Aliases {
	return &Aliases{ids: map[string]model.Id{}}
}

func (this *Aliases) Remember(alias string, id model.Id) {
	if alias != "" {
		this.ids[alias] = id
	}
}

func (this *Aliases) Id(alias string) (model.Id, bool) {
	id, found := this.ids[alias]
	return id, found
}

// MustId returns an alias's id or an error naming it, for a module wiring an edge by hand.
func (this *Aliases) MustId(alias string) (model.Id, error) {
	id, found := this.ids[alias]
	if !found {
		return "", errors.Errorf("defaults: alias %q was not created before it was referenced", alias)
	}
	return id, nil
}

// Resolve returns a copy of row ready to create in orgId: "$alias" stripped, every "@alias"
// replaced by its id, and org_id set. It also returns the row's own alias.
func (this *Aliases) Resolve(row dmodel.DynamicFields, orgId model.Id) (dmodel.DynamicFields, string, error) {
	resolved := make(dmodel.DynamicFields, len(row)+1)
	alias := ""
	for key, value := range row {
		if key == AliasKey {
			alias, _ = value.(string)
			continue
		}
		replaced, err := this.resolveValue(value)
		if err != nil {
			return nil, "", errors.Wrapf(err, "field %s", key)
		}
		resolved[key] = replaced
	}
	resolved[basemodel.FieldOrgId] = string(orgId)
	return resolved, alias, nil
}

func (this *Aliases) resolveValue(value any) (any, error) {
	switch typed := value.(type) {
	case string:
		if !strings.HasPrefix(typed, RefPrefix) {
			return typed, nil
		}
		id, err := this.MustId(strings.TrimPrefix(typed, RefPrefix))
		return string(id), err
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			replaced, err := this.resolveValue(item)
			if err != nil {
				return nil, err
			}
			out[i] = replaced
		}
		return out, nil
	default:
		return value, nil
	}
}

// CreateFn creates one resolved row and returns its id.
type CreateFn func(row dmodel.DynamicFields) (model.Id, error)

// CreateAll resolves and creates rows in file order, remembering each row's alias. A row may
// reference an alias created earlier in the same file.
func (this *Aliases) CreateAll(rows []dmodel.DynamicFields, orgId model.Id, create CreateFn) error {
	for i, row := range rows {
		resolved, alias, err := this.Resolve(row, orgId)
		if err != nil {
			return errors.Wrapf(err, "defaults: row %d", i)
		}
		id, err := create(resolved)
		if err != nil {
			return errors.Wrapf(err, "defaults: row %d (alias %q)", i, alias)
		}
		this.Remember(alias, id)
	}
	return nil
}
