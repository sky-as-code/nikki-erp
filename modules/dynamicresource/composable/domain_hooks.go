package composable

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
)

// The validator hooks are aliases of the corecrud hook types, instantiated at *DynamicEntity,
// the type argument the default domain service passes to corecrud.Create and corecrud.Update.
// Being aliases rather than distinct named types, they assign straight into the crud params.
//
// A hook that mutates the entity in place and returns the same pointer works, because the field
// map is shared. A hook wanting to *replace* the map must return a fresh NewDynamicEntityFrom(...).

// BeforeValidationFn may sanitize or enrich the model before schema validation.
type BeforeValidationFn = corecrud.BeforeValidationFn[*DynamicEntity]

// AfterValidationFn runs after successful schema validation.
type AfterValidationFn = corecrud.AfterValidationSuccessFn[*DynamicEntity]

// ValidateExtraFn performs validation the schema cannot express.
//
// It is the corecrud *update* hook shape, the wider of the two, so one type serves create,
// update and delete. On update and delete, foundModel is the stored record. On create there is
// no stored record and foundModel is nil, so a hook that reads it must nil-check first.
type ValidateExtraFn = corecrud.UpdateValidateExtraFn[*DynamicEntity]

// CreateOptions are what a derived domain service passes when it calls the embedded default's
// Create. They replace package engine's ModifyAction: instead of a definition the service reads
// at call time, the override hands the hooks over on each call.
type CreateOptions struct {
	BeforeValidation       BeforeValidationFn
	AfterValidationSuccess AfterValidationFn
	ValidateExtra          ValidateExtraFn
}

type UpdateOptions struct {
	BeforeValidation       BeforeValidationFn
	AfterValidationSuccess AfterValidationFn
	ValidateExtra          ValidateExtraFn
}

type DeleteOptions struct {
	// ValidateExtra receives the stored record as foundModel: the default service fetches it
	// before running the hook, because a delete guard's whole job is to inspect the row.
	ValidateExtra ValidateExtraFn
}

// ChainValidateExtra runs the given hooks in order and stops at the first one that reports a
// violation, so a guard placed first keeps a later rule from running against input it already
// refused. Nil hooks are skipped; a chain of nothing is nil, which corecrud treats as "no hook".
func ChainValidateExtra(fns ...ValidateExtraFn) ValidateExtraFn {
	live := make([]ValidateExtraFn, 0, len(fns))
	for _, fn := range fns {
		if fn != nil {
			live = append(live, fn)
		}
	}
	switch len(live) {
	case 0:
		return nil
	case 1:
		return live[0]
	}
	return func(ctx corectx.Context, inputModel *DynamicEntity, foundModel *DynamicEntity, vErrs *ft.ClientErrors) error {
		for _, fn := range live {
			before := vErrs.Count()
			if err := fn(ctx, inputModel, foundModel, vErrs); err != nil {
				return errors.Wrap(err, "ChainValidateExtra")
			}
			if vErrs.Count() > before {
				return nil
			}
		}
		return nil
	}
}

// RejectArchivedOnCreate refuses a create that carries is_archived, which is set through the
// /archived action and never as a create field. The schema validator drops it silently,
// answering 201 with a record whose visibility is not what the caller asked for, so it is
// reported instead. Nil for a schema without an is_archived field.
//
// cmd is the command as the client sent it. The check cannot be made against the entity the
// hook receives: corecrud calls InjectServiceFields before ValidateExtra, which applies the
// schema's own default for is_archived, so by then the field is present on every create and the
// guard would refuse them all.
func RejectArchivedOnCreate(schema *dmodel.ModelSchema, cmd CreateCommand) ValidateExtraFn {
	if schema == nil {
		return nil
	}
	if _, ok := schema.Field(basemodel.FieldIsArchived); !ok {
		return nil
	}
	if _, sent := cmd[basemodel.FieldIsArchived]; !sent {
		return nil
	}
	return func(_ corectx.Context, _ *DynamicEntity, _ *DynamicEntity, vErrs *ft.ClientErrors) error {
		vErrs.Append(*ft.NewBusinessViolation(basemodel.FieldIsArchived,
			"common.immutable_at_create",
			"is_archived cannot be set at create time; use the archived action instead"))
		return nil
	}
}

// withCreateGuard chains a guard ahead of whatever ValidateExtra a caller passes to Create.
// The onion installs it when RejectArchivedOnCreate is requested, so a module does not have
// to repeat the guard in every derived service.
//
// The guard is built per call rather than once, because it decides from the command the client
// sent: the entity the hook receives has already been through InjectServiceFields.
type createGuardService struct {
	CrudDomainService
	newGuard func(schema *dmodel.ModelSchema, cmd CreateCommand) ValidateExtraFn
}

func withCreateGuard(
	base CrudDomainService, newGuard func(schema *dmodel.ModelSchema, cmd CreateCommand) ValidateExtraFn,
) CrudDomainService {
	if newGuard == nil {
		return base
	}
	return &createGuardService{CrudDomainService: base, newGuard: newGuard}
}

func (this *createGuardService) Create(
	ctx corectx.Context, cmd CreateCommand, options ...CreateOptions,
) (*CreateResult, error) {
	opts := firstOrZero(options)
	opts.ValidateExtra = ChainValidateExtra(this.newGuard(this.Schema(), cmd), opts.ValidateExtra)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func firstOrZero[T any](options []T) T {
	var zero T
	if len(options) > 0 {
		return options[0]
	}
	return zero
}
