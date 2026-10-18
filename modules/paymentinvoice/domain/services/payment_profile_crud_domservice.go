package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	itProfile "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentprofile"
)

// The payment profile's CRUD, with the credentials encrypted on the way in and decrypted on the
// way out.
//
// A caller sends and receives "config", the readable credentials, while the only column is
// "encrypted_config". The two are swapped either side of the default CRUD rather than inside the
// repository, so a profile read by any other path (a database client, a backup, the payment flow's
// own repository reads) yields cipher text and nothing else.
//
// "config" is deliberately not a schema field: declaring it would give it a column. The REST layer
// therefore binds the two write bodies itself so the field arrives at all (payment_profile_rest.go).

func NewPaymentProfileCrudDomainService(
	base composable.CrudDomainService, profiles *PaymentProfileDomainService,
) itProfile.PaymentProfileCrudDomainService {
	return &PaymentProfileCrudDomainServiceImpl{CrudDomainService: base, profiles: profiles}
}

type PaymentProfileCrudDomainServiceImpl struct {
	composable.CrudDomainService
	profiles *PaymentProfileDomainService
}

// Create encrypts before the default runs: "config" is not a schema field, so validation would
// drop it and there would be nothing left to encrypt afterwards.
func (this *PaymentProfileCrudDomainServiceImpl) Create(
	ctx corectx.Context, cmd itProfile.CreatePaymentProfileCommand, options ...composable.CreateOptions,
) (*itProfile.CreatePaymentProfileResult, error) {
	if err := this.encryptParams(cmd); err != nil {
		return nil, err
	}
	result, err := this.CrudDomainService.Create(ctx, cmd, options...)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	// Give the caller back what they sent, not the cipher text.
	return result, this.profiles.DecryptFields(result.Data)
}

func (this *PaymentProfileCrudDomainServiceImpl) Update(
	ctx corectx.Context, cmd itProfile.UpdatePaymentProfileCommand, options ...composable.UpdateOptions,
) (*itProfile.UpdatePaymentProfileResult, error) {
	if err := this.encryptParams(cmd); err != nil {
		return nil, err
	}
	return this.CrudDomainService.Update(ctx, cmd, options...)
}

func (this *PaymentProfileCrudDomainServiceImpl) GetById(
	ctx corectx.Context, query itProfile.GetPaymentProfileByIdQuery,
) (*itProfile.GetPaymentProfileByIdResult, error) {
	swapConfigForEncryptedConfig(query)
	result, err := this.CrudDomainService.GetById(ctx, query)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	return result, this.decryptSingle(result.Data.Item, result.Data.Meta.DesiredFields)
}

func (this *PaymentProfileCrudDomainServiceImpl) GetOne(
	ctx corectx.Context, query composable.GetOneQuery,
) (*composable.GetOneResult, error) {
	swapConfigForEncryptedConfig(query)
	result, err := this.CrudDomainService.GetOne(ctx, query)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	return result, this.decryptSingle(result.Data.Item, result.Data.Meta.DesiredFields)
}

func (this *PaymentProfileCrudDomainServiceImpl) Search(
	ctx corectx.Context, query itProfile.SearchPaymentProfilesQuery, options ...corecrud.ServiceSearchOptions,
) (*itProfile.SearchPaymentProfilesResult, error) {
	swapConfigForEncryptedConfig(query)
	result, err := this.CrudDomainService.Search(ctx, query, options...)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	for _, item := range result.Data.Items {
		if err := this.profiles.DecryptFields(item); err != nil {
			return nil, err
		}
	}
	// The caller asked for "config"; echoing back the column name it resolved to would have the
	// frontend look for a key the payload does not contain.
	renameField(result.Data.DesiredFields,
		models.PaymentProfileFieldEncryptedConfig, models.PaymentProfileFieldConfig)
	return result, nil
}

func (this *PaymentProfileCrudDomainServiceImpl) decryptSingle(item dmodel.DynamicFields, desired []string) error {
	renameField(desired, models.PaymentProfileFieldEncryptedConfig, models.PaymentProfileFieldConfig)
	return this.profiles.DecryptFields(item)
}

// encryptParams swaps the plain credentials in the command for their encrypted form, in place.
func (this *PaymentProfileCrudDomainServiceImpl) encryptParams(params dmodel.DynamicFields) error {
	if _, hasConfig := params[models.PaymentProfileFieldConfig]; !hasConfig {
		return nil
	}
	return this.profiles.EncryptConfig(models.NewPaymentProfileFrom(params))
}

// swapConfigForEncryptedConfig rewrites a requested "config" field to the column that holds it.
// Without it a caller selecting the credentials is answered "no such field".
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
