package dynamicengines

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
)

// The payment profile's CRUD, with the credentials encrypted on the way in and decrypted on the
// way out.
//
// A profile is served by the engine's built-in CRUD like every other resource here; what is added
// is the one thing the schema cannot express. A caller sends and receives "config", the readable
// credentials, while the only column is "encrypted_config". The two are swapped either side of the
// built-in action rather than inside the repository, so a profile read by any other path — a
// database client, a backup, a join from another module — yields cipher text and nothing else.
//
// "config" is deliberately not a schema field. Declaring it would give it a column, which is the
// very thing being avoided; the cost is that the engine's own request binding drops it, so the two
// write actions bind their body themselves. See paymentProfileWriteHandler.

// definePaymentProfileActions installs the encryption on the built-in CRUD.
//
// Only the actions that carry a record are touched. delete and set_archived move no credentials,
// and exists answers with ids, so all three are left exactly as the engine defines them.
func definePaymentProfileActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		wrapPaymentProfileWrite(engine, drif.ActionCreate),
		wrapPaymentProfileWrite(engine, drif.ActionUpdate),
		wrapPaymentProfileRead(engine, drif.ActionGetById),
		wrapPaymentProfileRead(engine, drif.ActionGetByUnique),
		wrapPaymentProfileRead(engine, drif.ActionSearch),
	)
}

// wrapPaymentProfileWrite encrypts the credentials before the built-in action persists them, and
// installs the body binding that lets them arrive at all.
func wrapPaymentProfileWrite(engine drif.DynamicResourceEngine, actionName string) error {
	builtIn, err := builtInProcess(engine, actionName)
	if err != nil {
		return err
	}

	return engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:  actionName,
		RestHandler: paymentProfileWriteHandler(engine, actionName),
		MainProcess: func(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
			// Must run before the built-in action: "config" is not a schema field, so the crud
			// helper's validation drops it and there would be nothing left to encrypt afterwards.
			if err := encryptProfileParams(input.Params); err != nil {
				return nil, err
			}

			result, err := builtIn(ctx, input)
			if err != nil || result == nil || !result.HasData {
				return result, err
			}

			// Give the caller back what they sent us, not the cipher text.
			return result, decryptProfileResult(result)
		},
	})
}

// wrapPaymentProfileRead resolves a request for "config" against the column that actually holds it
// and decrypts whatever comes back.
func wrapPaymentProfileRead(engine drif.DynamicResourceEngine, actionName string) error {
	builtIn, err := builtInProcess(engine, actionName)
	if err != nil {
		return err
	}

	return engine.ModifyAction(drif.DynamicActionDelta{
		ActionName: actionName,
		MainProcess: func(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
			swapConfigForEncryptedConfig(input.Params)

			result, err := builtIn(ctx, input)
			if err != nil || result == nil || !result.HasData {
				return result, err
			}

			return result, decryptProfileResult(result)
		},
	})
}

// builtInProcess returns the action's current main process, so a wrapper can delegate to it.
//
// A missing action is a Go error rather than a silent skip: it means the engine was built without
// the built-in CRUD, and a profile resource whose create is not wrapped would write the plain
// credentials into the column.
func builtInProcess(
	engine drif.DynamicResourceEngine, actionName string,
) (drif.DynamicActionProcessFn, error) {
	definition, exists := engine.Action(actionName)
	if !exists {
		return nil, errors.Errorf(
			"the '%s' resource engine has no built-in '%s' action to wrap",
			models.PaymentProfileSchemaName, actionName,
		)
	}

	return definition.MainProcess, nil
}

// encryptProfileParams swaps the plain credentials in the request for their encrypted form,
// in place, so the params the built-in action goes on to persist no longer carry them.
func encryptProfileParams(params dmodel.DynamicFields) error {
	if _, hasConfig := params[models.PaymentProfileFieldConfig]; !hasConfig {
		return nil
	}

	service, err := requirePaymentProfileService()
	if err != nil {
		return err
	}

	return service.EncryptConfig(models.NewPaymentProfileFrom(params))
}

// decryptProfileResult fills in the plain credentials on whatever shape the action answered with,
// and drops the cipher text.
//
// Every branch mutates a map or a slice the result still points at, so reassigning result.Data is
// unnecessary. A shape not listed here carries no record — delete and exists answer with counts
// and ids — and is left alone.
func decryptProfileResult(result *drif.ActionResult) error {
	service, err := requirePaymentProfileService()
	if err != nil {
		return err
	}

	switch data := result.Data.(type) {
	case dmodel.DynamicFields:
		return service.DecryptFields(data)

	case dyn.SingleResultData[dmodel.DynamicFields]:
		return service.DecryptFields(data.Item)

	case dyn.PagedResultData[dmodel.DynamicFields]:
		for _, item := range data.Items {
			if err := service.DecryptFields(item); err != nil {
				return err
			}
		}
		// The caller asked for "config"; echoing back the column name we resolved it to would
		// have the frontend look for a key the payload does not contain.
		renameField(data.DesiredFields,
			models.PaymentProfileFieldEncryptedConfig, models.PaymentProfileFieldConfig)
		return nil
	}

	return nil
}

// swapConfigForEncryptedConfig rewrites a requested "config" field to the column that holds it.
// Without it a caller selecting the credentials is answered "no such field": "config" is not on
// the schema, precisely because it is not stored.
func swapConfigForEncryptedConfig(params dmodel.DynamicFields) {
	if params == nil {
		return
	}

	switch fields := params[basemodel.FieldFields].(type) {
	case []string:
		renameField(fields, models.PaymentProfileFieldConfig, models.PaymentProfileFieldEncryptedConfig)
	case []any:
		for index, field := range fields {
			if name, isString := field.(string); isString && name == models.PaymentProfileFieldConfig {
				fields[index] = models.PaymentProfileFieldEncryptedConfig
			}
		}
	}
}

// renameField replaces every occurrence of one field name in a selection, in place.
func renameField(fields []string, from string, to string) {
	for index, field := range fields {
		if field == from {
			fields[index] = to
		}
	}
}

// paymentProfileService is the domain service the profile actions encrypt and decrypt through. It
// is a package variable rather than a container lookup for the same reason the order and invoice
// services are: an action callback is handed only its own engine.
var paymentProfileService *services.PaymentProfileDomainService

// SetPaymentProfileService installs the service the profile actions delegate to. Init calls it
// before any request is served.
func SetPaymentProfileService(service *services.PaymentProfileDomainService) {
	paymentProfileService = service
}

func requirePaymentProfileService() (*services.PaymentProfileDomainService, error) {
	if paymentProfileService == nil {
		return nil, errors.New(
			"the payment profile domain service was not installed; PaymentInvoiceModule.Init must " +
				"call dynamicengines.SetPaymentProfileService")
	}

	return paymentProfileService, nil
}
