package models

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
	PasswordStoreSchemaName = "authenticate_password_store"

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
		Label(model.NewLangJsonRefSf("%s.label", PasswordStoreSchemaName)).
		TableName("authn_password_stores").
		ShouldBuildDb().
		CompositeUnique(PasswordStoreFieldPrincipalType, PasswordStoreFieldPrincipalId).
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			DefinePrincipalTypeField(PasswordStoreFieldPrincipalType).
				Label(model.NewLangJsonRefSf("fields.%s", PasswordStoreFieldPrincipalType)).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(PasswordStoreFieldPrincipalId).
				Label(model.NewLangJsonRefSf("fields.%s", PasswordStoreFieldPrincipalId)).
				RequiredForCreate(),
		).
		Field(
			DefinePasswordTextField(PasswordStoreFieldPassword).
				Label(model.NewLangJsonRefSf("fields.%s", PasswordStoreFieldPassword)),
		).
		Field(
			dmodel.DefineField().Name(PasswordStoreFieldPasswordExpiresAt).
				Label(model.NewLangJsonRefSf("fields.%s", PasswordStoreFieldPasswordExpiresAt)).
				DataType(dmodel.FieldDataTypeDateTime()),
		).
		Field(
			dmodel.DefineField().Name(PasswordStoreFieldPasswordUpdatedAt).
				Label(model.NewLangJsonRefSf("fields.%s", PasswordStoreFieldPasswordUpdatedAt)).
				DataType(dmodel.FieldDataTypeDateTime()),
		).
		Field(
			DefinePasswordTextField(PasswordStoreFieldPasswordTmp).
				Label(model.NewLangJsonRefSf("fields.%s", PasswordStoreFieldPasswordTmp)),
		).
		Field(
			dmodel.DefineField().Name(PasswordStoreFieldPasswordTmpExpiresAt).
				Label(model.NewLangJsonRefSf("fields.%s", PasswordStoreFieldPasswordTmpExpiresAt)).
				DataType(dmodel.FieldDataTypeDateTime()),
		).
		Field(
			DefinePasswordOtpField(PasswordStoreFieldPasswordOtp).
				Label(model.NewLangJsonRefSf("fields.%s", PasswordStoreFieldPasswordOtp)),
		).
		Field(
			dmodel.DefineField().Name(PasswordStoreFieldPasswordOtpExpiresAt).
				Label(model.NewLangJsonRefSf("fields.%s", PasswordStoreFieldPasswordOtpExpiresAt)).
				DataType(dmodel.FieldDataTypeDateTime()),
		).
		Field(
			DefinePasswordOtpRecoveryField(PasswordStoreFieldPasswordOtpRecovery).
				Label(model.NewLangJsonRefSf("fields.%s", PasswordStoreFieldPasswordOtpRecovery)),
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
	basemodel.DynamicModelBase
}

func NewPasswordStore() *PasswordStore {
	return &PasswordStore{basemodel.NewDynamicModel()}
}

func NewPasswordStoreFrom(src dmodel.DynamicFields) *PasswordStore {
	return &PasswordStore{basemodel.NewDynamicModel(src)}
}

func (this PasswordStore) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *PasswordStore) SetId(v *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, v)
}

func (this PasswordStore) GetPassword() *string {
	return this.GetFieldData().GetString(PasswordStoreFieldPassword)
}

func (this *PasswordStore) SetPassword(v *string) {
	this.GetFieldData().SetString(PasswordStoreFieldPassword, v)
}

func (this PasswordStore) GetPasswordExpiresAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(PasswordStoreFieldPasswordExpiresAt)
}

func (this *PasswordStore) SetPasswordExpiresAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(PasswordStoreFieldPasswordExpiresAt, v)
}

func (this PasswordStore) GetPasswordUpdatedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(PasswordStoreFieldPasswordUpdatedAt)
}

func (this *PasswordStore) SetPasswordUpdatedAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(PasswordStoreFieldPasswordUpdatedAt, v)
}

func (this PasswordStore) GetPasswordTmp() *string {
	return this.GetFieldData().GetString(PasswordStoreFieldPasswordTmp)
}

func (this *PasswordStore) SetPasswordTmp(v *string) {
	this.GetFieldData().SetString(PasswordStoreFieldPasswordTmp, v)
}

func (this PasswordStore) GetPasswordTmpExpiresAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(PasswordStoreFieldPasswordTmpExpiresAt)
}

func (this *PasswordStore) SetPasswordTmpExpiresAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(PasswordStoreFieldPasswordTmpExpiresAt, v)
}

func (this PasswordStore) MustGetPasswordOtp() string {
	v := this.GetFieldData().GetString(PasswordStoreFieldPasswordOtp)
	if v == nil {
		panic("password OTP is not set")
	}
	return *v
}

func (this PasswordStore) GetPasswordOtp() *string {
	return this.GetFieldData().GetString(PasswordStoreFieldPasswordOtp)
}

func (this *PasswordStore) SetPasswordOtp(v *string) {
	this.GetFieldData().SetString(PasswordStoreFieldPasswordOtp, v)
}

func (this PasswordStore) GetPasswordOtpExpiresAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(PasswordStoreFieldPasswordOtpExpiresAt)
}

func (this *PasswordStore) SetPasswordOtpExpiresAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(PasswordStoreFieldPasswordOtpExpiresAt, v)
}

func (this PasswordStore) GetPasswordOtpRecovery() []string {
	return this.GetFieldData().GetStrings(PasswordStoreFieldPasswordOtpRecovery)
}

func (this *PasswordStore) SetPasswordOtpRecovery(v []string) {
	this.GetFieldData().SetStrings(PasswordStoreFieldPasswordOtpRecovery, v)
}

func (this PasswordStore) GetPrincipalType() *PrincipalType {
	s := this.GetFieldData().GetString(PasswordStoreFieldPrincipalType)
	if s == nil {
		return nil
	}
	st := PrincipalType(*s)
	return &st
}

func (this *PasswordStore) SetPrincipalType(v *PrincipalType) {
	if v == nil {
		this.GetFieldData().SetString(PasswordStoreFieldPrincipalType, nil)
		return
	}
	this.GetFieldData().SetString(PasswordStoreFieldPrincipalType, (*string)(v))
}

func (this PasswordStore) GetPrincipalId() *model.Id {
	return this.GetFieldData().GetModelId(PasswordStoreFieldPrincipalId)
}

func (this *PasswordStore) SetPrincipalId(v *model.Id) {
	this.GetFieldData().SetModelId(PasswordStoreFieldPrincipalId, v)
}
