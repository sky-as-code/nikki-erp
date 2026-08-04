package iam

import (
	"go.bryk.io/pkg/errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

// initDynamicEngines creates the resource engines this module owns and publishes them
// into the dependency container, so that other modules can inject them by name.
//
// User is the pilot resource of the dynamic resource engine. Its hand-written layers
// (application service, domain service, repository and REST handlers) are untouched and
// keep serving /v1/iam/users; the engine serves the same resource at /v1/iam/iam_user.
func initDynamicEngines() error {
	userEngine, err := dynamicresource.Registry().NewEngine(models.UserSchemaName)
	if err != nil {
		return errors.Wrap(err, "failed to create the user resource engine")
	}

	if err := defineUserActions(userEngine); err != nil {
		return err
	}

	return deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.UserSchemaName),
		func() drif.DynamicResourceEngine { return userEngine },
	)
}

// defineUserActions adds the user-specific actions on top of the built-in CRUD ones.
func defineUserActions(userEngine drif.DynamicResourceEngine) error {
	err := userEngine.DefineAction(drif.DynamicActionDefinition{
		ActionName:  "send_invitation",
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
