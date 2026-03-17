package domain

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
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
				val.Length(1, model.MODEL_RULE_URL_LENGTH),
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
	userFieldAvatarUrl   = "avatar_url"
	userFieldDisplayName = "display_name"
	userFieldEmail       = "email"
	userFieldStatus      = "status"
)

func UserSchemaBuilder() *schema.EntitySchemaBuilder {
	return schema.DefineEntity("identity.user").
		Label(model.LangJson{"en-US": "User"}).
		TableName("ident_users").
		CompositeUnique("email", "display_name").
		Field(
			schema.DefineField().
				Name("id").
				Label(model.LangJson{"en-US": "ID"}).
				DataType(schema.FieldDataTypeModelId()).
				PrimaryKey(),
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
				Name(userFieldDisplayName).
				Label(model.LangJson{"en-US": "Display Name"}).
				DataType(schema.FieldDataTypeString()).
				Required().
				Rule(schema.FieldRuleLength(1, model.MODEL_RULE_LONG_NAME_LENGTH)),
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
				Name("group_id").
				Label(model.LangJson{"en-US": "ID"}).
				DataType(schema.FieldDataTypeModelId()).
				Foreign(schema.Edge("group").ManyToOne("identity.group", "id")),
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
