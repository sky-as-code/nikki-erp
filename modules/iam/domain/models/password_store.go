package models

import (
	"math"
	"regexp"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
)

type OtpCode string

var OtpCodePattern = regexp.MustCompile(`^\d+$`)

type PasswordStoreType string

const (
	PasswordStoreTypePassword     = PasswordStoreType("password")
	PasswordStoreTypePasswordTemp = PasswordStoreType("passwordtmp")
	PasswordStoreTypeOtpSecret    = PasswordStoreType("otp_secret")
	PasswordStoreTypeOtpRecovery  = PasswordStoreType("otp_recovery")
)

const (
	PasswordStoreSchemaName = "authenticate_password_store"

	PasswordStoreFieldId            = basemodel.FieldId
	PasswordStoreFieldPrincipalType = "principal_type"
	PasswordStoreFieldPrincipalId   = "principal_id"
	PasswordStoreFieldType          = "type"
	PasswordStoreFieldHash          = "hash"
	PasswordStoreFieldExpiresAt     = "expires_at"
	PasswordStoreFieldLastUsedAt    = "last_used_at"
)

func PasswordStoreSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(PasswordStoreSchemaName).
		TableName("iam_password_stores").
		ShouldBuildDb().
		CompositeUnique(dmodel.CompositeUniqueParam{
			IndexName: "iam_pwd_stores_tid_princ_type_princ_id_type",
			Fields:    []string{PasswordStoreFieldPrincipalType, PasswordStoreFieldPrincipalId, PasswordStoreFieldType},
		}).
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			DefinePrincipalTypeField(PasswordStoreFieldPrincipalType).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(PasswordStoreFieldPrincipalId).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(PasswordStoreFieldType).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(PasswordStoreTypePassword), string(PasswordStoreTypePasswordTemp), string(PasswordStoreTypeOtpSecret), string(PasswordStoreTypeOtpRecovery),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(PasswordStoreFieldHash).
				DataType(dmodel.FieldDataTypeSecret(0, math.MaxInt16)),
		).
		Field(
			dmodel.DefineField().Name(PasswordStoreFieldLastUsedAt).
				DataType(dmodel.FieldDataTypeDateTime()),
		).
		Field(
			dmodel.DefineField().Name(PasswordStoreFieldExpiresAt).
				DataType(dmodel.FieldDataTypeDateTime()),
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

func (this PasswordStore) GetType() *PasswordStoreType {
	s := this.GetFieldData().GetString(PasswordStoreFieldType)
	if s == nil {
		return nil
	}
	st := PasswordStoreType(*s)
	return &st
}

func (this *PasswordStore) SetType(v *PasswordStoreType) {
	if v == nil {
		this.GetFieldData().SetString(PasswordStoreFieldType, nil)
		return
	}
	this.GetFieldData().SetString(PasswordStoreFieldType, (*string)(v))
}

func (this PasswordStore) MustGetHash() string {
	v := this.GetHash()
	if v == nil {
		panic("hash is nil")
	}
	return *v
}

func (this PasswordStore) GetHash() *string {
	return this.GetFieldData().GetString(PasswordStoreFieldHash)
}

func (this *PasswordStore) SetHash(v *string) {
	this.GetFieldData().SetString(PasswordStoreFieldHash, v)
}

func (this PasswordStore) GetLastUsedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(PasswordStoreFieldLastUsedAt)
}

func (this *PasswordStore) SetLastUsedAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(PasswordStoreFieldLastUsedAt, v)
}

func (this PasswordStore) GetExpiresAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(PasswordStoreFieldExpiresAt)
}

func (this *PasswordStore) SetExpiresAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(PasswordStoreFieldExpiresAt, v)
}
