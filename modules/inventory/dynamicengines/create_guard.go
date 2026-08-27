package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// rejectArchivedOnCreate refuses a create that carries is_archived.
//
// is_archived comes from core.basemodel.archivable_model and is set through the /archived action,
// never as a create field. The schema validator drops it silently, which answers 201 with a record
// whose visibility is not what the caller asked for; a caller creating an already-invisible record
// is telling us the request is wrong, so it is reported rather than ignored.
//
// It wraps the action's existing ValidateExtra instead of replacing it, because ModifyAction
// overwrites that field and every guard the specs installed must keep running.
func rejectArchivedOnCreate(engine drif.DynamicResourceEngine) error {
	if _, ok := engine.Schema().Field(basemodel.FieldIsArchived); !ok {
		return nil
	}

	definition, ok := engine.Action(drif.ActionCreate)
	if !ok {
		return nil
	}
	existing := definition.ValidateExtra

	return engine.ModifyAction(drif.DynamicActionDelta{
		ActionName: drif.ActionCreate,
		ValidateExtra: func(
			ctx corectx.Context, inputModel *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
		) error {
			if _, sent := inputModel.GetFieldData()[basemodel.FieldIsArchived]; sent {
				vErrs.Append(*ft.NewBusinessViolation(basemodel.FieldIsArchived,
					"common.immutable_at_create",
					"is_archived cannot be set at create time; use the archived action instead"))
				return nil
			}
			if existing == nil {
				return nil
			}
			return errors.Wrap(existing(ctx, inputModel, foundModel, vErrs), "rejectArchivedOnCreate")
		},
	})
}
