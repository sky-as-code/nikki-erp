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
		DefaultFields: []string{
			models.UserFieldAvatarUrl,
			models.UserFieldDisplayName,
			models.UserFieldEmail,
			models.UserFieldStatus,
		},
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

// processSendUserInvitation shows how a custom action reaches the record the engine
// already fetched for it. Sending the actual email is left to the invitation feature.
func processSendUserInvitation(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	found, err := input.ResourceRepository.FindByKeys(ctx, userKeysToFetch(input.Params))
	if err != nil {
		return nil, errors.Wrap(err, "processSendUserInvitation")
	}
	if !found.HasData {
		return &drif.ActionResult{HasData: false}, nil
	}

	return &drif.ActionResult{
		Data:    dmodel.DynamicFields{models.UserFieldEmail: found.Data[models.UserFieldEmail]},
		HasData: true,
	}, nil
}
