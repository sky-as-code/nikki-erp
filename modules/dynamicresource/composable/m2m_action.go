package composable

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
)

// A many-to-many action links and unlinks the peers of one record: POST {schema}/:id/manage_<peers>
// with {"add": [ids], "remove": [ids]}. The domain service runs it through ManageM2m, which owns
// the transaction and the hooks; the application service authorizes it first with
// AssertRecordAction, because the permission is often not the resource's own (assigning roles to a
// group asserts manage_role_assignments, not update).

// ManageM2mCommand carries "id" (the source record, from the path), "add" and "remove".
type ManageM2mCommand = dmodel.DynamicFields

const (
	FieldM2mAdd    = "add"
	FieldM2mRemove = "remove"
)

// M2mChange is what one manage call asked to link and unlink, handed to the hooks.
type M2mChange struct {
	SrcId   model.Id
	Added   []model.Id
	Removed []model.Id
}

// AllPeerIds lists the added and the removed ids together, for a hook that refreshes every peer
// the call touched.
func (this M2mChange) AllPeerIds() []model.Id {
	return append(append([]model.Id{}, this.Added...), this.Removed...)
}

// ManageM2mParam declares one many-to-many action of a resource.
type ManageM2mParam struct {
	// Action labels the operation in error messages ("manage group users").
	Action string

	// DestSchemaName is the peer schema; the source schema must declare a many-to-many relation
	// to it.
	DestSchemaName string

	// JunctionIdField names the junction table's own id column, filled with a fresh id on every
	// inserted row. Empty for a junction keyed by the pair alone.
	JunctionIdField string

	// Validate runs before anything is written, inside the transaction. A violation appended to
	// vErrs refuses the call.
	Validate func(ctx corectx.Context, change M2mChange, vErrs *ft.ClientErrors) error

	// AfterChange runs after the junction rows changed, inside the same transaction, so a
	// permission rebuild or an audit row stands or falls with the link. It does not run when the
	// call changed nothing. A returned error rolls the whole call back.
	AfterChange func(tranxCtx corectx.Context, change M2mChange) error
}

// m2mLinkWriterFn writes the junction rows. It is crud.ManageM2m in production and a stub in the
// unit tests, which have no database.
type m2mLinkWriterFn func(
	ctx corectx.Context, repo CrudRepository, change M2mChange, param ManageM2mParam,
) (*MutateResult, error)

// ManageM2m links and unlinks the peers named in cmd, in one transaction with the hooks.
//
// A call that changes nothing answers a result without data, which the REST layer reports as
// not-found exactly like the hand-written manage routes it replaces.
func ManageM2m(
	ctx corectx.Context, repo CrudRepository, cmd ManageM2mCommand, param ManageM2mParam,
) (*MutateResult, error) {
	return manageM2mWith(ctx, repo, cmd, param, writeM2mLinks)
}

func manageM2mWith(
	ctx corectx.Context, repo CrudRepository, cmd ManageM2mCommand, param ManageM2mParam,
	writeLinks m2mLinkWriterFn,
) (*MutateResult, error) {
	change, cErrs := ParseM2mChange(cmd)
	if cErrs != nil {
		return &MutateResult{ClientErrors: *cErrs}, nil
	}

	var result *MutateResult
	refused, err := InRefusableTransaction(ctx, repo, func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
		vErrs := ft.ClientErrors{}
		if param.Validate != nil {
			if err := param.Validate(tranxCtx, *change, &vErrs); err != nil {
				return nil, err
			}
			if vErrs.Count() > 0 {
				return &vErrs, nil
			}
		}
		var err error
		result, err = writeLinks(tranxCtx, repo, *change, param)
		if err != nil {
			return nil, err
		}
		if result.ClientErrors.Count() > 0 {
			return &result.ClientErrors, nil
		}
		if !result.HasData || param.AfterChange == nil {
			return nil, nil
		}
		return nil, param.AfterChange(tranxCtx, *change)
	})
	if err != nil {
		return nil, errors.Wrap(err, param.Action)
	}
	if refused != nil {
		return &MutateResult{ClientErrors: *refused}, nil
	}
	return result, nil
}

// ParseM2mChange reads the id and the add/remove id lists of a manage command. A list that is not
// an array of strings is the caller's mistake and answers a malformed-request error.
func ParseM2mChange(cmd ManageM2mCommand) (*M2mChange, *ft.ClientErrors) {
	var lists struct {
		Add    datastructure.Set[model.Id] `json:"add"`
		Remove datastructure.Set[model.Id] `json:"remove"`
	}
	payload := dmodel.DynamicFields{}
	for _, field := range []string{FieldM2mAdd, FieldM2mRemove} {
		if val, ok := cmd[field]; ok && val != nil {
			payload[field] = val
		}
	}
	if err := decodeParams(payload, &lists); err != nil {
		return nil, &ft.ClientErrors{*ft.NewValidationError(FieldM2mAdd, ft.ErrorKey("err_malformed_request"),
			"add and remove must be arrays of ids")}
	}
	change := &M2mChange{SrcId: ParamId(cmd)}
	if lists.Add != nil {
		change.Added = lists.Add.ToSlice()
	}
	if lists.Remove != nil {
		change.Removed = lists.Remove.ToSlice()
	}
	return change, nil
}

func writeM2mLinks(
	ctx corectx.Context, repo CrudRepository, change M2mChange, param ManageM2mParam,
) (*MutateResult, error) {
	return corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:             param.Action,
		DbRepoGetter:       repo,
		DestSchemaName:     param.DestSchemaName,
		SrcId:              change.SrcId,
		SrcIdFieldForError: "id",
		AssociatedIds:      datastructure.NewSetFrom(change.Added...),
		DisassociatedIds:   datastructure.NewSetFrom(change.Removed...),
		BeforeInsert:       junctionIdFiller(param.JunctionIdField),
	})
}

func junctionIdFiller(field string) func(ctx corectx.Context, dbRecords []dmodel.DynamicFields) error {
	if field == "" {
		return nil
	}
	return func(_ corectx.Context, dbRecords []dmodel.DynamicFields) error {
		for _, record := range dbRecords {
			newId, err := model.NewId()
			if err != nil {
				return errors.Wrap(err, "junctionIdFiller")
			}
			record[field] = *newId
		}
		return nil
	}
}
