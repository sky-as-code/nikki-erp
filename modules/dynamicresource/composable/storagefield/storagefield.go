// Package storagefield is the rule shared by every resource that keeps `<name>_file_path` columns
// pointing into IT Storage. The client uploads to /v1/itstorage/files/upload first, then writes
// the returned path through the resource's ordinary JSON create or update; this rule checks that
// the path names a file the record's org may use, and removes the file a write replaced or cleared.
//
// A domain service chains ValidateCreate into its Create and wraps its Update in Guard.Update.
package storagefield

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// Port is what the rule needs of IT Storage. Each owning module declares its own port in
// interfaces/external with these two methods, so it satisfies this one structurally.
type Port interface {
	AssertFileUsable(ctx corectx.Context, path string, orgId *model.Id) (*model.Id, error)
	DeleteFileByPath(ctx corectx.Context, path string) error
}

// Scope says whose space a path must lie in.
type Scope int

const (
	// ScopeOrg confines a path to the record's org_id space.
	ScopeOrg Scope = iota
	// ScopeTenant is for a resource without org_id (iam_user): any file of the tenant.
	ScopeTenant
)

type Guard struct {
	port   Port
	scope  Scope
	fields []string
}

func NewGuard(port Port, scope Scope, fields ...string) Guard {
	return Guard{port: port, scope: scope, fields: fields}
}

// Fields are the `*_file_path` columns the guard watches.
func (this Guard) Fields() []string {
	return this.fields
}

// ValidateCreate checks every path the create carries. A create replaces nothing, so it never
// deletes.
func (this Guard) ValidateCreate() composable.ValidateExtraFn {
	return func(ctx corectx.Context, input, _ *composable.DynamicEntity, vErrs *ft.ClientErrors) error {
		return this.validate(ctx, input.GetFieldData(), nil, vErrs, nil)
	}
}

// Update runs the update with the guard chained last into its ValidateExtra, then removes every
// path the update replaced or cleared. The removal is best-effort and happens only once the write
// is durable: inside a caller's transaction it is skipped, because a rollback would leave the
// record pointing at a deleted file, which is worse than an orphaned one.
func (this Guard) Update(
	ctx corectx.Context, opts composable.UpdateOptions,
	run func(opts composable.UpdateOptions) (*composable.MutateResult, error),
) (*composable.MutateResult, error) {
	replaced := []string{}
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra,
		func(ctx corectx.Context, input, found *composable.DynamicEntity, vErrs *ft.ClientErrors) error {
			return this.validate(ctx, input.GetFieldData(), fieldsOf(found), vErrs, &replaced)
		})

	result, err := run(opts)
	if err != nil || result == nil || result.ClientErrors.Count() > 0 || !result.HasData {
		return result, err
	}
	this.deleteReplaced(ctx, replaced)
	return result, nil
}

func (this Guard) validate(
	ctx corectx.Context, input, found dmodel.DynamicFields, vErrs *ft.ClientErrors, replaced *[]string,
) error {
	orgId := this.orgOf(input, found)
	for _, field := range this.fields {
		if _, sent := input[field]; !sent {
			continue
		}
		newPath, oldPath := pathOf(input, field), pathOf(found, field)
		if newPath == oldPath {
			continue
		}
		if newPath != "" {
			if err := this.assertUsable(ctx, field, newPath, orgId, vErrs); err != nil {
				return err
			}
		}
		if oldPath != "" && replaced != nil {
			*replaced = append(*replaced, oldPath)
		}
	}
	return nil
}

func (this Guard) assertUsable(
	ctx corectx.Context, field, path string, orgId *model.Id, vErrs *ft.ClientErrors,
) error {
	fileId, err := this.port.AssertFileUsable(ctx, path, orgId)
	if err != nil {
		if items, isClient := clientErrorsOf(err); isClient {
			for _, item := range items {
				item.Field = field
				vErrs.Append(item)
			}
			return nil
		}
		return errors.Wrapf(err, "failed to check the file of '%s'", field)
	}
	if fileId == nil {
		vErrs.Append(*ft.NewBusinessViolation(field, ft.ErrorKey("file.err_path_not_usable"),
			"'"+path+"' is not a file this record may use"))
	}
	return nil
}

func (this Guard) deleteReplaced(ctx corectx.Context, paths []string) {
	if len(paths) == 0 || ctx.GetDbTranx() != nil {
		return
	}
	for _, path := range paths {
		if err := this.port.DeleteFileByPath(ctx, path); err != nil && ctx.GetLogger() != nil {
			ctx.GetLogger().Warnf("failed to delete the replaced file '%s': %v", path, err)
		}
	}
}

// orgOf prefers the stored org: an update need not repeat org_id, and a record never moves org.
func (this Guard) orgOf(input, found dmodel.DynamicFields) *model.Id {
	if this.scope == ScopeTenant {
		return nil
	}
	for _, fields := range []dmodel.DynamicFields{found, input} {
		if orgId := pathOf(fields, basemodel.FieldOrgId); orgId != "" {
			id := model.Id(orgId)
			return &id
		}
	}
	return nil
}

// clientErrorsOf recognizes the shapes a port may report a refusal in, so a bad path answers 400
// rather than 500.
func clientErrorsOf(err error) (ft.ClientErrors, bool) {
	if signal, ok := composable.AsClientErrorSignal(err); ok {
		return signal.Errors, true
	}
	var item *ft.ClientErrorItem
	if errors.As(err, &item) && item != nil {
		return ft.ClientErrors{*item}, true
	}
	return nil, false
}

func fieldsOf(entity *composable.DynamicEntity) dmodel.DynamicFields {
	if entity == nil {
		return nil
	}
	return entity.GetFieldData()
}

func pathOf(fields dmodel.DynamicFields, field string) string {
	if fields == nil {
		return ""
	}
	switch value := fields[field].(type) {
	case string:
		return value
	case *string:
		if value != nil {
			return *value
		}
	}
	return ""
}
