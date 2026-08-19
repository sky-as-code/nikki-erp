package models

import (
	_ "embed"
	"encoding/json"
	"strconv"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// PaymentProfileMethod is the gateway a profile holds credentials for.
//
// The first three values are byte-identical to the AdapterCode constants in payment_method.go on
// purpose: a profile is resolved to an adapter by this value alone, and a translation table
// between the two would be one more place for them to drift apart.
type PaymentProfileMethod string

const (
	PaymentProfileMethodMomo   = PaymentProfileMethod(AdapterCodeMomo)
	PaymentProfileMethodVietQr = PaymentProfileMethod(AdapterCodeVietQr)
	PaymentProfileMethodMpos   = PaymentProfileMethod(AdapterCodeMpos)

	// PaymentProfileMethodMbbank is accepted by the schema because the service this module
	// supersedes stored profiles under it, and a profile that cannot be written back is a
	// profile that cannot be migrated. No adapter answers to it yet.
	PaymentProfileMethodMbbank = PaymentProfileMethod("mbbank")
)

const (
	PaymentProfileSchemaName = "paymentinvoice_payment_profile"

	PaymentProfileFieldId     = basemodel.FieldId
	PaymentProfileFieldName   = "name"
	PaymentProfileFieldMethod = "method"
	PaymentProfileFieldOrgId  = "org_id"

	// PaymentProfileFieldConfig holds the plain, readable credentials. It is deliberately not a
	// schema field and therefore has no database column: it only carries data in from requests
	// and out to responses, and is swapped for PaymentProfileFieldEncryptedConfig before a write
	// reaches the repository.
	PaymentProfileFieldConfig = "config"

	// PaymentProfileFieldEncryptedConfig holds the encrypted form of PaymentProfileFieldConfig
	// and is the only one of the two that is persisted.
	PaymentProfileFieldEncryptedConfig = "encrypted_config"
)

//go:embed payment_profile.json
var paymentProfileSchemaJson string

func PaymentProfileSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(paymentProfileSchemaJson)
}

// PaymentProfile is one merchant account at one gateway: the credentials an order is collected
// with, held as a row rather than as configuration.
//
// It is the counterpart of coremart's vdmc_payments table, and carries the same columns, because
// the two are written by the same operators against the same gateway contracts. What it is not is
// a second PaymentMethod: a method says what a payer may choose and what amounts it accepts, a
// profile says which merchant account the money lands in. One deployment routinely has several
// profiles for one method — a second MoMo partner code for a different store — which is why the
// method column carries no unique constraint.
//
// Credentials never sit in the clear: only encrypted_config is a column, and the plain config
// exists solely between the request and the encryption step, or between the decryption step and
// the response.
type PaymentProfile struct {
	basemodel.DynamicModelBase
}

func NewPaymentProfile() *PaymentProfile {
	return &PaymentProfile{basemodel.NewDynamicModel()}
}

func NewPaymentProfileFrom(src dmodel.DynamicFields) *PaymentProfile {
	return &PaymentProfile{basemodel.NewDynamicModel(src)}
}

func (this PaymentProfile) GetName() *string {
	return this.GetFieldData().GetString(PaymentProfileFieldName)
}

func (this *PaymentProfile) SetName(v *string) {
	this.GetFieldData().SetString(PaymentProfileFieldName, v)
}

func (this PaymentProfile) GetMethod() *PaymentProfileMethod {
	raw := this.GetFieldData().GetString(PaymentProfileFieldMethod)
	if raw == nil {
		return nil
	}
	method := PaymentProfileMethod(*raw)

	return &method
}

func (this *PaymentProfile) SetMethod(v *PaymentProfileMethod) {
	if v == nil {
		this.GetFieldData().SetString(PaymentProfileFieldMethod, nil)
		return
	}
	raw := string(*v)
	this.GetFieldData().SetString(PaymentProfileFieldMethod, &raw)
}

// GetConfig returns the plain credentials. It is reached through GetAny because the field is not
// declared on the schema and so has no typed accessor; only the gateway adapter named by the
// method interprets what is in it.
//
// It answers nil on a record read straight from the repository: the config is filled in by the
// decryption step, which the CRUD read actions and PaymentProfileDomainService both run.
func (this PaymentProfile) GetConfig() map[string]any {
	raw := this.GetFieldData().GetAny(PaymentProfileFieldConfig)
	if raw == nil {
		return nil
	}

	config, ok := raw.(map[string]any)
	if !ok {
		return nil
	}

	return config
}

func (this *PaymentProfile) SetConfig(v map[string]any) {
	this.GetFieldData().SetAny(PaymentProfileFieldConfig, v)
}

func (this PaymentProfile) GetEncryptedConfig() *string {
	return this.GetFieldData().GetString(PaymentProfileFieldEncryptedConfig)
}

func (this *PaymentProfile) SetEncryptedConfig(v *string) {
	this.GetFieldData().SetString(PaymentProfileFieldEncryptedConfig, v)
}

// ParseConfigToEncryptedConfig serializes the plain "config" field to JSON, hands it to encryptFn
// and stores the result in "encrypted_config". The plain "config" field is then dropped so that it
// never reaches the database.
//
// It does nothing when the model carries no "config" key at all, which is the case for a partial
// update that does not touch the credentials: writing an absent field as null would wipe the
// credentials of every profile edited for its name alone. A "config" key present but null is a
// different statement — the caller is clearing it — and is honoured.
func (this *PaymentProfile) ParseConfigToEncryptedConfig(encryptFn func(plain string) (string, error)) error {
	fields := this.GetFieldData()
	if _, hasConfig := fields[PaymentProfileFieldConfig]; !hasConfig {
		return nil
	}
	defer delete(fields, PaymentProfileFieldConfig)

	config := fields.GetAny(PaymentProfileFieldConfig)
	if config == nil {
		this.SetEncryptedConfig(nil)
		return nil
	}

	plain, err := json.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "parse payment profile config to encrypted config")
	}

	encrypted, err := encryptFn(string(plain))
	if err != nil {
		return errors.Wrap(err, "parse payment profile config to encrypted config")
	}
	this.SetEncryptedConfig(&encrypted)

	return nil
}

// ParseEncryptedConfigToConfig hands the stored "encrypted_config" field to decryptFn and parses
// the result back into the plain "config" field. The "encrypted_config" field is then dropped so
// that the cipher text is never exposed to callers.
//
// A record fetched without the encrypted_config column carries no key for it, and is left exactly
// as it is: a listing that selected only name and method must not fail because it did not ask for
// the credentials.
func (this *PaymentProfile) ParseEncryptedConfigToConfig(decryptFn func(encrypted string) (string, error)) error {
	fields := this.GetFieldData()
	if _, hasEncrypted := fields[PaymentProfileFieldEncryptedConfig]; !hasEncrypted {
		return nil
	}
	defer delete(fields, PaymentProfileFieldEncryptedConfig)

	encrypted := this.GetEncryptedConfig()
	if encrypted == nil || len(*encrypted) == 0 {
		return nil
	}

	plain, err := decryptFn(*encrypted)
	if err != nil {
		return errors.Wrap(err, "parse payment profile encrypted config to config")
	}

	var config any
	if err := json.Unmarshal([]byte(plain), &config); err != nil {
		return errors.Wrap(err, "parse payment profile encrypted config to config")
	}
	fields.SetAny(PaymentProfileFieldConfig, config)

	return nil
}

// ConfigValues returns the credentials as a flat map of name to value.
//
// The operator console stores a profile's config as a list of {key, value} entries rather than as
// a plain object, because it renders them as an editable table — so a profile written there
// decrypts to {"0": {"key": "partnerCode", "value": "..."}, ...} rather than to
// {"partnerCode": "..."}. Flattening it here is what lets every adapter read a credential by name
// without knowing how the console chose to store it.
//
// A config already written as a plain object is returned as it stands, so a profile created by an
// import or by hand works too; an entry that is neither shape is kept under the key it was filed
// under rather than dropped, so nothing an operator typed disappears silently.
func (this PaymentProfile) ConfigValues() map[string]any {
	entries := this.configEntries()
	if entries == nil {
		return nil
	}

	values := make(map[string]any, len(entries))
	for name, raw := range entries {
		key, value, isEntry := readConfigEntry(raw)
		if isEntry {
			values[key] = value
			continue
		}
		values[name] = raw
	}

	return values
}

// configEntries reads the decrypted config into a name-keyed map, whichever of the two JSON shapes
// it arrived in. A list is indexed by position, which ConfigValues then discards in favour of each
// entry's own key.
func (this PaymentProfile) configEntries() map[string]any {
	raw := this.GetFieldData().GetAny(PaymentProfileFieldConfig)
	if raw == nil {
		return nil
	}

	switch config := raw.(type) {
	case map[string]any:
		return config
	case []any:
		entries := make(map[string]any, len(config))
		for index, entry := range config {
			entries[strconv.Itoa(index)] = entry
		}
		return entries
	}

	return nil
}

// readConfigEntry recognizes one {key, value} entry, reporting false for anything else.
func readConfigEntry(raw any) (key string, value any, isEntry bool) {
	entry, isObject := raw.(map[string]any)
	if !isObject {
		return "", nil, false
	}

	name, hasName := entry["key"].(string)
	if !hasName || name == "" {
		return "", nil, false
	}
	stored, hasValue := entry["value"]
	if !hasValue {
		return "", nil, false
	}

	return name, stored, true
}
