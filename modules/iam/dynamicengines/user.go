package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

func userEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.UserSchemaName,
		DefineActions: defineUserActions,
	}
}

// defineUserActions adds the user-specific actions on top of the built-in CRUD ones.
func defineUserActions(userEngine drif.DynamicResourceEngine) error {
	err := userEngine.DefineAction(drif.DynamicActionDefinition{
		ActionName:  "send_invitation",
		ActionType:  drif.ActionTypeGeneric,
		RestPath:    ":id/send_invitation",
		Permission:  drif.PermissionUpdate,
		KeysToFetch: userKeysToFetch,
		MainProcess: processSendUserInvitation,
	})
	return errors.Wrap(err, "failed to define user actions")
}

func userKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{models.UserFieldId: params[models.UserFieldId]}
}

// processSendUserInvitation shows how a custom action reaches the record the engine already
// fetched for it: KeysToFetch identified the user, and the pipeline hands it over as FoundModel,
// so there is no second read. Sending the actual email is left to the invitation feature.
func processSendUserInvitation(
	_ corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	if input.FoundModel == nil {
		return &drif.ActionResult{HasData: false}, nil
	}

	return &drif.ActionResult{
		Data:    dmodel.DynamicFields{models.UserFieldEmail: (*input.FoundModel)[models.UserFieldEmail]},
		HasData: true,
	}, nil
}
