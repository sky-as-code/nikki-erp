package services

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/crypto"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// paymentProfilePageSize bounds a sweep over the profiles of one gateway.
//
// A deployment has a handful of merchant accounts per gateway, not hundreds. Reading a fixed page
// rather than paging exhaustively keeps a lookup that runs on every inbound callback to one query;
// a deployment that genuinely exceeds this has outgrown the assumption and should be told so
// rather than silently searched half way.
const paymentProfilePageSize = 200

// NewPaymentProfileDomainService reads the key the profiles' credentials are encrypted with.
//
// It does not fail when the key is unset. This module is linked into both binaries, and only one
// of them configures the key today; refusing to construct would take the whole application down at
// boot over a resource a given deployment may never use. The failure is deferred to the first
// profile that actually needs encrypting or decrypting, where it is reported against that request,
// and a warning is logged at boot so it is not first discovered there.
func NewPaymentProfileDomainService(
	configSvc config.ConfigService, logger logging.LoggerService,
) *PaymentProfileDomainService {
	encryptionKey := configSvc.GetStr(constants.PaymentProfileEncryptionKey)
	if len(encryptionKey) == 0 && logger != nil {
		logger.Warnf(
			"paymentinvoice: %s is unset; payment profile credentials can be neither stored nor read",
			string(constants.PaymentProfileEncryptionKey),
		)
	}

	return &PaymentProfileDomainService{encryptionKey: encryptionKey}
}

// PaymentProfileDomainService owns the one rule a payment profile has: its credentials are
// readable in a request and unreadable in the database.
//
// Everything else about a profile is plain CRUD and is served by the resource engine. This service
// exists so that the engine's actions and the payment flow encrypt and decrypt through the same
// code — the flow reads a profile on every payment, and a second implementation of the same
// unwrapping is a second place for the format to drift.
type PaymentProfileDomainService struct {
	encryptionKey string
}

// EncryptConfig moves the plain config into the encrypted, persisted field.
// Must be called on every write, right before the record reaches the database.
func (this *PaymentProfileDomainService) EncryptConfig(profile *models.PaymentProfile) error {
	if err := this.assertKeyConfigured(); err != nil {
		return err
	}

	return profile.ParseConfigToEncryptedConfig(func(plain string) (string, error) {
		return crypto.EncryptString(plain, this.encryptionKey)
	})
}

// DecryptConfig moves the persisted encrypted config back into the plain config field.
// Must be called on every read, right after the record leaves the database.
func (this *PaymentProfileDomainService) DecryptConfig(profile *models.PaymentProfile) error {
	// A record fetched without the encrypted_config column has nothing to decrypt, and a listing
	// that never asked for the credentials must not be refused for want of a key it does not need.
	if _, hasEncrypted := profile.GetFieldData()[models.PaymentProfileFieldEncryptedConfig]; !hasEncrypted {
		return nil
	}
	if err := this.assertKeyConfigured(); err != nil {
		return err
	}

	return profile.ParseEncryptedConfigToConfig(func(encrypted string) (string, error) {
		return crypto.DecryptString(encrypted, this.encryptionKey)
	})
}

// DecryptFields is DecryptConfig over a field map, for the callers that hold one rather than a
// model: the engine's read actions hand back DynamicFields, and rewrapping each of them at every
// call site would put the same three lines in four places.
func (this *PaymentProfileDomainService) DecryptFields(fields dmodel.DynamicFields) error {
	if fields == nil {
		return nil
	}
	profile := models.NewPaymentProfileFrom(fields)

	return this.DecryptConfig(profile)
}

// FindById fetches one profile by primary key, with its credentials decrypted.
//
// It answers (nil, nil) when no such profile exists, because a caller quoting an id that names
// nothing has made a mistake the caller can fix, not one this module failed at.
func (this *PaymentProfileDomainService) FindById(
	ctx corectx.Context, profileId string,
) (*models.PaymentProfile, error) {
	engine, err := engineFor(models.PaymentProfileSchemaName)
	if err != nil {
		return nil, err
	}

	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.PaymentProfileFieldId: profileId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "PaymentProfileDomainService.FindById")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	profile := models.NewPaymentProfileFrom(found.Data)
	if err := this.DecryptConfig(profile); err != nil {
		return nil, errors.Wrap(err, "PaymentProfileDomainService.FindById")
	}

	return profile, nil
}

// FindActiveByMethod returns every profile of one gateway that has not been archived, credentials
// decrypted.
//
// It serves the callbacks that identify themselves by a merchant account rather than by an order —
// a card terminal posts its merchant id and an encrypted body, so the profile whose credentials
// decrypt that body has to be found before anything in it can be read. Archived profiles are left
// out: an account withdrawn from use must not be able to settle new payments, and including it
// would make archiving a label rather than a control.
func (this *PaymentProfileDomainService) FindActiveByMethod(
	ctx corectx.Context, method models.PaymentProfileMethod,
) ([]*models.PaymentProfile, error) {
	engine, err := engineFor(models.PaymentProfileSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.PaymentProfileFieldMethod, dmodel.Equals, string(method)),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  paymentProfilePageSize,
	})
	if err != nil {
		return nil, errors.Wrap(err, "PaymentProfileDomainService.FindActiveByMethod")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	profiles := make([]*models.PaymentProfile, 0, len(found.Data.Items))
	for _, item := range found.Data.Items {
		profile := models.NewPaymentProfileFrom(item)
		if err := this.DecryptConfig(profile); err != nil {
			return nil, errors.Wrap(err, "PaymentProfileDomainService.FindActiveByMethod")
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (this *PaymentProfileDomainService) assertKeyConfigured() error {
	if len(this.encryptionKey) > 0 {
		return nil
	}

	return errors.Errorf(
		"missing config '%s'; payment profile credentials cannot be encrypted or decrypted",
		string(constants.PaymentProfileEncryptionKey),
	)
}
