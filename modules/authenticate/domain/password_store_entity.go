package domain

import (
	"regexp"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	c "github.com/sky-as-code/nikki-erp/modules/authenticate/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type OtpCode string

var OtpCodePattern = regexp.MustCompile(`^\d+$`)

const (
	PasswordStoreSchemaName = "authenticate.password_store"

	PasswordStoreFieldId                   = basemodel.FieldId
	PasswordStoreFieldPrincipalType        = "principal_type"
	PasswordStoreFieldPrincipalId          = "principal_id"
	PasswordStoreFieldPassword             = "password"
	PasswordStoreFieldPasswordExpiresAt    = "password_expires_at"
	PasswordStoreFieldPasswordUpdatedAt    = "password_updated_at"
	PasswordStoreFieldPasswordTmp          = "passwordtmp"
	PasswordStoreFieldPasswordTmpExpiresAt = "passwordtmp_expires_at"
	PasswordStoreFieldPasswordOtp          = "passwordotp"
	PasswordStoreFieldPasswordOtpExpiresAt = "passwordotp_expires_at"
	PasswordStoreFieldPasswordOtpRecovery  = "passwordotp_recovery"
)

func PasswordStoreSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(PasswordStoreSchemaName).
		Label(model.LangJson{"en-US": "Password Store"}).
		TableName("authn_password_stores").
		ShouldBuildDb().
		CompositeUnique(PasswordStoreFieldPrincipalType, PasswordStoreFieldPrincipalId).
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			DefinePrincipalTypeField(PasswordStoreFieldPrincipalType).
				Label(model.LangJson{"en-US": "Principal type"}).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(PasswordStoreFieldPrincipalId).
				Label(model.LangJson{"en-US": "Principal ID"}).
				RequiredForCreate(),
		).
		Field(
			DefinePasswordTextField(PasswordStoreFieldPassword).
				Label(model.LangJson{"en-US": "Password"}),
		).
		Field(
			dmodel.DefineField().Name(PasswordStoreFieldPasswordExpiresAt).
				Label(model.LangJson{"en-US": "Password expired at"}).
				DataType(dmodel.FieldDataTypeDateTime()),
		).
		Field(
			dmodel.DefineField().Name(PasswordStoreFieldPasswordUpdatedAt).
				Label(model.LangJson{"en-US": "Password updated at"}).
				DataType(dmodel.FieldDataTypeDateTime()),
		).
		Field(
			DefinePasswordTextField(PasswordStoreFieldPasswordTmp).
				Label(model.LangJson{"en-US": "Temporary password"}),
		).
		Field(
			dmodel.DefineField().Name(PasswordStoreFieldPasswordTmpExpiresAt).
				Label(model.LangJson{"en-US": "Temporary password expired at"}).
				DataType(dmodel.FieldDataTypeDateTime()),
		).
		Field(
			DefinePasswordOtpField(PasswordStoreFieldPasswordOtp).
				Label(model.LangJson{"en-US": "OTP secret"}),
		).
		Field(
			dmodel.DefineField().Name(PasswordStoreFieldPasswordOtpExpiresAt).
				Label(model.LangJson{"en-US": "OTP expired at"}).
				DataType(dmodel.FieldDataTypeDateTime()),
		).
		Field(
			DefinePasswordOtpRecoveryField(PasswordStoreFieldPasswordOtpRecovery).
				Label(model.LangJson{"en-US": "OTP recovery"}),
		)
}

func DefinePasswordTextField(fieldName string) *dmodel.FieldBuilder {
	return dmodel.DefineField().Name(fieldName).
		DataType(dmodel.FieldDataTypeSecret(model.MODEL_RULE_PASSWORD_MIN_LENGTH, model.MODEL_RULE_PASSWORD_MAX_LENGTH))
}

func DefinePasswordOtpField(fieldName string) *dmodel.FieldBuilder {
	return dmodel.DefineField().Name(fieldName).
		DataType(dmodel.FieldDataTypeSecret(c.OtpCodeLength, c.OtpCodeLength))
}

func DefinePasswordOtpRecoveryField(fieldName string) *dmodel.FieldBuilder {
	return dmodel.DefineField().Name(fieldName).
		DataType(dmodel.FieldDataTypeSecret(c.OtpRecoveryCodeLength, c.OtpRecoveryCodeLength).ArrayType())
}

type PasswordStore struct {
	fields dmodel.DynamicFields
}

func NewPasswordStore() *PasswordStore {
	return &PasswordStore{fields: make(dmodel.DynamicFields)}
}

func NewPasswordStoreFrom(src dmodel.DynamicFields) *PasswordStore {
	return &PasswordStore{fields: src}
}

func (this PasswordStore) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *PasswordStore) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this PasswordStore) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *PasswordStore) SetId(v *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, v)
}

func (this PasswordStore) GetPassword() *string {
	return this.fields.GetString(PasswordStoreFieldPassword)
}

func (this *PasswordStore) SetPassword(v *string) {
	this.fields.SetString(PasswordStoreFieldPassword, v)
}

func (this PasswordStore) GetPasswordExpiresAt() *model.ModelDateTime {
	return this.fields.GetModelDateTime(PasswordStoreFieldPasswordExpiresAt)
}

func (this *PasswordStore) SetPasswordExpiresAt(v *model.ModelDateTime) {
	this.fields.SetModelDateTime(PasswordStoreFieldPasswordExpiresAt, v)
}

func (this PasswordStore) GetPasswordUpdatedAt() *model.ModelDateTime {
	return this.fields.GetModelDateTime(PasswordStoreFieldPasswordUpdatedAt)
}

func (this *PasswordStore) SetPasswordUpdatedAt(v *model.ModelDateTime) {
	this.fields.SetModelDateTime(PasswordStoreFieldPasswordUpdatedAt, v)
}

func (this PasswordStore) GetPasswordTmp() *string {
	return this.fields.GetString(PasswordStoreFieldPasswordTmp)
}

func (this *PasswordStore) SetPasswordTmp(v *string) {
	this.fields.SetString(PasswordStoreFieldPasswordTmp, v)
}

func (this PasswordStore) GetPasswordTmpExpiresAt() *model.ModelDateTime {
	return this.fields.GetModelDateTime(PasswordStoreFieldPasswordTmpExpiresAt)
}

func (this *PasswordStore) SetPasswordTmpExpiresAt(v *model.ModelDateTime) {
	this.fields.SetModelDateTime(PasswordStoreFieldPasswordTmpExpiresAt, v)
}

func (this PasswordStore) MustGetPasswordOtp() string {
	v := this.fields.GetString(PasswordStoreFieldPasswordOtp)
	if v == nil {
		panic("password OTP is not set")
	}
	return *v
}

func (this PasswordStore) GetPasswordOtp() *string {
	return this.fields.GetString(PasswordStoreFieldPasswordOtp)
}

func (this *PasswordStore) SetPasswordOtp(v *string) {
	this.fields.SetString(PasswordStoreFieldPasswordOtp, v)
}

func (this PasswordStore) GetPasswordOtpExpiresAt() *model.ModelDateTime {
	return this.fields.GetModelDateTime(PasswordStoreFieldPasswordOtpExpiresAt)
}

func (this *PasswordStore) SetPasswordOtpExpiresAt(v *model.ModelDateTime) {
	this.fields.SetModelDateTime(PasswordStoreFieldPasswordOtpExpiresAt, v)
}

func (this PasswordStore) GetPasswordOtpRecovery() []string {
	return this.fields.GetStrings(PasswordStoreFieldPasswordOtpRecovery)
}

func (this *PasswordStore) SetPasswordOtpRecovery(v []string) {
	this.fields.SetStrings(PasswordStoreFieldPasswordOtpRecovery, v)
}

func (this PasswordStore) GetPrincipalType() *PrincipalType {
	s := this.fields.GetString(PasswordStoreFieldPrincipalType)
	if s == nil {
		return nil
	}
	st := PrincipalType(*s)
	return &st
}

func (this *PasswordStore) SetPrincipalType(v *PrincipalType) {
	if v == nil {
		this.fields.SetString(PasswordStoreFieldPrincipalType, nil)
		return
	}
	this.fields.SetString(PasswordStoreFieldPrincipalType, (*string)(v))
}

func (this PasswordStore) GetPrincipalId() *model.Id {
	return this.fields.GetModelId(PasswordStoreFieldPrincipalId)
}

func (this *PasswordStore) SetPrincipalId(v *model.Id) {
	this.fields.SetModelId(PasswordStoreFieldPrincipalId, v)
}
