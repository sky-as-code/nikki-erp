package services

import (
	"strings"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

// principalTypeScope confines a domain service over iam_users to one principal type. iam_user and
// iam_service share the table and the user repository; which rows each one sees is decided here,
// not in the repository, so the unfiltered reads (permission evaluation, the request guard) keep
// going straight to the repository.
//
// Every search is ANDed with `type = X`, every id-addressed action first checks the row is of
// type X (a row of the other type answers not-found, as if absent), and a create writes X.
type principalTypeScope struct {
	principalType domain.UserType
}

func (this principalTypeScope) typeNode() *dmodel.SearchNode {
	return dmodel.NewSearchNode().NewCondition(domain.UserFieldType, dmodel.Equals, string(this.principalType))
}

// narrowSearch ANDs the type condition after the caller's graph, keeping the caller's order.
func (this principalTypeScope) narrowSearch(params composable.SearchQuery) error {
	graph, err := dmodel.AndDecodedGraph(params[graphParam], *this.typeNode())
	if err != nil {
		return err
	}
	params[graphParam] = graph
	return nil
}

// isOfType reports whether the record params address is of this scope's type. A params map
// without an id, or with an id the repository rejects, is left to the default service, which
// reports the missing or malformed key itself.
func (this principalTypeScope) isOfType(
	ctx corectx.Context, repo composable.CrudRepository, params dmodel.DynamicFields,
) (bool, error) {
	id := composable.ParamId(params)
	if id == "" || !isWellFormedId(repo, id) {
		return true, nil
	}
	matched, invalid, err := this.match(ctx, repo, id)
	return matched || invalid, err
}

func isWellFormedId(repo composable.CrudRepository, id model.Id) bool {
	schema := repo.Schema()
	if schema == nil {
		return true
	}
	field, ok := schema.Field(basemodel.FieldId)
	if !ok {
		return true
	}
	_, cErr := field.DataType().Validate(dmodel.Value(string(id)))
	return cErr == nil
}

// match looks the principal up by id and type; invalid reports an id the repository refused.
func (this principalTypeScope) match(
	ctx corectx.Context, repo composable.CrudRepository, id model.Id,
) (matched bool, invalid bool, err error) {
	found, err := repo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{basemodel.FieldId: string(id), domain.UserFieldType: string(this.principalType)},
		Fields: []string{basemodel.FieldId},
	})
	if err != nil {
		return false, false, err
	}
	if found.ClientErrors.Count() > 0 {
		return false, true, nil
	}
	return found.HasData, false, nil
}

// assertOfType is isOfType for a mutating action: a mismatch is reported as not-found.
func (this principalTypeScope) assertOfType(
	ctx corectx.Context, repo composable.CrudRepository, params dmodel.DynamicFields,
) (*composable.MutateResult, error) {
	ok, err := this.isOfType(ctx, repo, params)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, nil
	}
	return &composable.MutateResult{ClientErrors: ft.ClientErrors{*ft.NewAnonymousNotFoundError()}}, nil
}

// refuseIdsOfType refuses every id in ids that is a principal of this scope's type: a membership or
// an assignment the other kind of principal holds alone. An id naming no principal at all is left
// to the foreign key, which reports it as before.
func (this principalTypeScope) refuseIdsOfType(
	ctx corectx.Context, repo composable.CrudRepository, field string, ids []model.Id, vErrs *ft.ClientErrors,
) error {
	for _, id := range ids {
		matched, _, err := this.match(ctx, repo, id)
		if err != nil {
			return err
		}
		if matched {
			vErrs.Append(*this.refusedError(field, id))
		}
	}
	return nil
}

func (this principalTypeScope) refusedError(field string, id model.Id) *ft.ClientErrorItem {
	if this.principalType == domain.UserTypeService {
		return ft.NewBusinessViolation(field, ft.ErrorKey(errPrincipalNotUser, "iam"),
			"a service cannot be given this; only a user can", map[string]any{"id": string(id)})
	}
	return ft.NewBusinessViolation(field, ft.ErrorKey(errPrincipalNotService, "iam"),
		"a user cannot be given this; only a service can", map[string]any{"id": string(id)})
}

// refuseServiceIn is a ValidateExtra refusing a service in the id field named field, for an owner
// that must be a person.
func refuseServiceIn(repo composable.CrudRepository, field string) composable.ValidateExtraFn {
	return func(ctx corectx.Context, input, _ *composable.DynamicEntity, vErrs *ft.ClientErrors) error {
		if input == nil {
			return nil
		}
		id, ok := input.GetFieldData()[field].(string)
		if !ok || id == "" {
			return nil
		}
		return serviceTypeScope.refuseIdsOfType(ctx, repo, field, []model.Id{model.Id(id)}, vErrs)
	}
}

const (
	graphParam             = "graph"
	fieldsParam            = "fields"
	errPrincipalNotUser    = "err_principal_not_user"
	errPrincipalNotService = "err_principal_not_service"
)

var (
	userTypeScope    = principalTypeScope{principalType: domain.UserTypeUser}
	serviceTypeScope = principalTypeScope{principalType: domain.UserTypeService}
)

func readableFieldNames(schema *dmodel.ModelSchema) []string {
	fields := schema.ReadableFields()
	names := make([]string, 0, len(fields))
	for _, field := range fields {
		names = append(names, field.Name())
	}
	return names
}

// restrictFields drops from fields= every path whose root is not a field of schema: a schema
// viewing a shared table must not project the columns it leaves out, which the shared
// repository would otherwise happily read.
func restrictFields(schema *dmodel.ModelSchema, params dmodel.DynamicFields) {
	var names []string
	switch raw := params[fieldsParam].(type) {
	case []string:
		names = raw
	case []any:
		for _, item := range raw {
			if name, ok := item.(string); ok {
				names = append(names, name)
			}
		}
	default:
		return
	}
	kept := make([]string, 0, len(names))
	for _, name := range names {
		root, _, _ := strings.Cut(name, ".")
		if _, ok := schema.Field(root); ok || root == "*" {
			kept = append(kept, name)
		}
	}
	if len(kept) == 0 {
		kept = schema.PrimaryKeys()
	}
	params[fieldsParam] = kept
}
