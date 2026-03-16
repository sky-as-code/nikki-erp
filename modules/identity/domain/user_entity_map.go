package domain

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"github.com/sky-as-code/nikki-erp/common/model"
)

const (
	userFieldDisplayName = "display_name"
	userFieldAvatarUrl   = "avatar_url"
	userFieldEmail       = "email"
	userFieldStatus      = "status"
	userFieldHierarchyId = "hierarchy_id"
)

func UserSchemaBuilder() *schema.EntitySchemaBuilder {
	return schema.DefineEntity().
		Label(model.LangJson{"en-US": "User"}).
		TableName("ident_users").
		Field(
			schema.DefineField().
				Name(userFieldDisplayName).
				Label(model.LangJson{"en-US": "Display Name"}).
				DataType(schema.FieldDataTypeString()).
				Required().
				Rule(schema.FieldRuleLength(1, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			schema.DefineField().
				Name(userFieldAvatarUrl).
				Label(model.LangJson{"en-US": "Avatar URL"}).
				DataType(schema.FieldDataTypeUrl()).
				Rule(schema.FieldRuleLength(1, model.MODEL_RULE_URL_LENGTH)),
		).
		Field(
			schema.DefineField().
				Name(userFieldEmail).
				Label(model.LangJson{"en-US": "Email"}).
				DataType(schema.FieldDataTypeEmail()).
				Required().
				Unique().
				Rule(schema.FieldRuleLength(5, model.MODEL_RULE_USERNAME_LENGTH)),
		).
		Field(
			schema.DefineField().
				Name(userFieldStatus).
				Label(model.LangJson{"en-US": "Status"}).
				DataType(schema.FieldDataTypeEnumString([]string{
					string(UserStatusActive), string(UserStatusArchived), string(UserStatusLocked),
				})).
				Required().
				Rule(schema.FieldRuleOneOf(UserStatusActive, UserStatusArchived, UserStatusLocked)),
		).
		Field(
			schema.DefineField().
				Name(userFieldHierarchyId).
				DataType(schema.FieldDataTypeUlid()),
		)
}

type UserEntity struct {
	fields schema.DynamicEntity
}

func NewUserEntity() *UserEntity {
	return &UserEntity{fields: make(schema.DynamicEntity)}
}

func NewUserEntityFrom(src schema.DynamicEntity) *UserEntity {
	return &UserEntity{fields: src}
}

func (this *UserEntity) GetFieldData() schema.DynamicEntity {
	return this.fields
}

func (this *UserEntity) SetFieldData(data schema.DynamicEntity) {
	this.fields = data
}

func (this *UserEntity) GetDisplayName() *string {
	return this.fields.GetString(userFieldDisplayName)
}

func (this *UserEntity) SetDisplayName(v *string) {
	this.fields.SetString(userFieldDisplayName, v)
}

func (this *UserEntity) GetAvatarUrl() *string {
	return this.fields.GetString(userFieldAvatarUrl)
}

func (this *UserEntity) SetAvatarUrl(v *string) {
	this.fields.SetString(userFieldAvatarUrl, v)
}

func (this *UserEntity) GetEmail() *string {
	return this.fields.GetString(userFieldEmail)
}

func (this *UserEntity) SetEmail(v *string) {
	this.fields.SetString(userFieldEmail, v)
}

func (this *UserEntity) GetStatus() *UserStatus {
	s := this.fields.GetString(userFieldStatus)
	if s == nil {
		return nil
	}
	st := UserStatus(*s)
	return &st
}

func (this *UserEntity) SetStatus(v *UserStatus) {
	if v == nil {
		this.fields.SetString(userFieldStatus, nil)
		return
	}
	s := string(*v)
	this.fields.SetString(userFieldStatus, &s)
}

func (this *UserEntity) GetHierarchyId() *model.Id {
	return this.fields.GetModelId(userFieldHierarchyId)
}

func (this *UserEntity) SetHierarchyId(v *model.Id) {
	this.fields.SetModelId(userFieldHierarchyId, v)
}
