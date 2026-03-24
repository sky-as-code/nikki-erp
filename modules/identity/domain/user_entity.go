package domain

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicentity/basemodel"
)

type User struct {
	model.ModelBase
	model.AuditableBase

	AvatarUrl   *string     `json:"avatarUrl"`
	DisplayName *string     `json:"displayName"`
	Email       *string     `json:"email"`
	HierarchyId *model.Id   `json:"hierarchyId"`
	OrgId       *model.Id   `json:"orgId"`
	Status      *UserStatus `json:"status,omitempty"`
	ScopeRef    *model.Id   `json:"scopeRef,omitempty" model:"-"`

	Groups    []Group          `json:"groups,omitempty" model:"-"` // TODO: Handle copy
	Hierarchy []HierarchyLevel `json:"hierarchy,omitempty" model:"-"`
	Orgs      []Organization   `json:"orgs,omitempty" model:"-"`
}

func (this *User) SetDefaults() {
	this.ModelBase.SetDefaults()
}

func (this *User) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.AvatarUrl,
			val.When(this.AvatarUrl != nil,
				val.Length(1, model.MODEL_RULE_URL_LENGTH_MAX),
				val.IsUrl,
			),
		),
		val.Field(&this.DisplayName,
			val.NotNilWhen(!forEdit),
			val.When(this.DisplayName != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.Email,
			val.NotNilWhen(!forEdit),
			val.When(this.Email != nil,
				val.NotEmpty,
				val.IsEmail,
				val.Length(5, model.MODEL_RULE_USERNAME_LENGTH),
			),
		),
		UserStatusValidateRule(&this.Status),
		model.IdPtrValidateRule(&this.HierarchyId, false),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type UserStatus string

const (
	UserStatusActive   = UserStatus("active")
	UserStatusArchived = UserStatus("archived")
	UserStatusLocked   = UserStatus("locked")
)

func (this UserStatus) String() string {
	return string(this)
}

func WrapUserStatus(s string) *UserStatus {
	st := UserStatus(s)
	return &st
}

func UserStatusValidateRule(field **UserStatus) *val.FieldRules {
	return val.Field(field,
		val.When(*field != nil,
			val.NotEmpty,
			val.OneOf(UserStatusActive, UserStatusArchived, UserStatusLocked),
		),
	)
}

const (
	UserSchemaName       = "identity.user"
	UserFieldAvatarUrl   = "avatar_url"
	UserFieldDisplayName = "display_name"
	UserFieldEmail       = "email"
	UserFieldIsOwner     = "is_owner"
	UserFieldStatus      = "status"
)

func UserSchemaBuilder() *schema.EntitySchemaBuilder {
	return schema.DefineEntity(UserSchemaName).
		Label(model.LangJson{"en-US": "User"}).
		TableName("ident_users").
		CompositeUnique("email", "display_name").
		Extend(basemodel.BaseModelSchemaBuilder()).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Field(
			schema.DefineField().
				Name(UserFieldAvatarUrl).
				Label(model.LangJson{"en-US": "Avatar URL"}).
				DataType(schema.FieldDataTypeUrl()),
		).
		Field(
			schema.DefineField().
				Name(UserFieldDisplayName).
				Label(model.LangJson{"en-US": "Display Name"}).
				DataType(schema.FieldDataTypeString(3, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			schema.DefineField().
				Name(UserFieldEmail).
				Label(model.LangJson{"en-US": "Email"}).
				DataType(schema.FieldDataTypeEmail()).
				RequiredForCreate().
				Unique(),
		).
		Field(
			schema.DefineField().
				Name(UserFieldStatus).
				Label(model.LangJson{"en-US": "Status"}).
				DataType(schema.FieldDataTypeEnumString([]string{
					string(UserStatusActive), string(UserStatusLocked),
				})).
				RequiredForCreate().
				Default(string(UserStatusActive)),
		).
		Field(
			schema.DefineField().
				Name(UserFieldIsOwner).
				Label(model.LangJson{"en-US": "Is Owner"}).
				DataType(schema.FieldDataTypeBoolean()).
				Default(nil),
		).
		Field(
			schema.DefineField().
				Name("group_id").
				DataType(schema.FieldDataTypeModelId()).
				Foreign(
					schema.Edge("group").
						Label(model.LangJson{"en-US": "Group"}).
						ManyToOne("identity.group", "id").
						OnDelete(schema.RelationCascadeSetNull),
				),
		)

}

type UserEntity struct {
	fields schema.DynamicFields
}

func NewUserEntity() *UserEntity {
	return &UserEntity{fields: make(schema.DynamicFields)}
}

func NewUserEntityFrom(src schema.DynamicFields) *UserEntity {
	return &UserEntity{fields: src}
}

func (this UserEntity) GetFieldData() schema.DynamicFields {
	return this.fields
}

func (this *UserEntity) SetFieldData(data schema.DynamicFields) {
	this.fields = data
}

func (this UserEntity) GetDisplayName() *string {
	return this.fields.GetString(UserFieldDisplayName)
}

func (this *UserEntity) SetDisplayName(v *string) {
	this.fields.SetString(UserFieldDisplayName, v)
}

func (this UserEntity) GetAvatarUrl() *string {
	return this.fields.GetString(UserFieldAvatarUrl)
}

func (this *UserEntity) SetAvatarUrl(v *string) {
	this.fields.SetString(UserFieldAvatarUrl, v)
}

func (this UserEntity) GetEmail() *string {
	return this.fields.GetString(UserFieldEmail)
}

func (this *UserEntity) SetEmail(v *string) {
	this.fields.SetString(UserFieldEmail, v)
}

func (this UserEntity) IsOwner() bool {
	val := this.fields.GetBool(UserFieldIsOwner)
	if val == nil {
		return false
	}
	return *val
}

func (this *UserEntity) SetIsOwner(v *bool) {
	this.fields.SetBool(UserFieldIsOwner, v)
}

func (this UserEntity) GetStatus() *UserStatus {
	s := this.fields.GetString(UserFieldStatus)
	if s == nil {
		return nil
	}
	st := UserStatus(*s)
	return &st
}

func (this *UserEntity) SetStatus(v *UserStatus) {
	if v == nil {
		this.fields.SetString(UserFieldStatus, nil)
		return
	}
	s := string(*v)
	this.fields.SetString(UserFieldStatus, &s)
}
