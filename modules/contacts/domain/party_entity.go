package domain

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	it "github.com/sky-as-code/nikki-erp/modules/core/tag/interfaces"
)

type PartyTag = it.Tag

const (
	PartyTagType = "contacts_party"
)

type Party struct {
	model.ModelBase
	model.AuditableBase

	AvatarUrl    *string   `json:"avatarUrl,omitempty"`
	DisplayName  *string   `json:"displayName,omitempty"`
	LegalName    *string   `json:"legalName,omitempty"`
	LegalAddress *string   `json:"legalAddress,omitempty"`
	TaxId        *string   `json:"taxId,omitempty"`
	JobPosition  *string   `json:"jobPosition,omitempty"`
	Title        *string   `json:"title,omitempty"`
	Type         *string   `json:"type,omitempty"`
	Note         *string   `json:"note,omitempty"`
	Nationality  *model.Id `json:"nationality,omitempty"`
	OrgId        *model.Id `json:"orgId,omitempty"`
	Language     *model.Id `json:"language,omitempty"`
	Website      *string   `json:"website,omitempty"`

	// Relations
	Tags          []PartyTag     `json:"tags,omitempty" model:"-"`
	CommChannels  []CommChannel  `json:"commChannels,omitempty" model:"-"`
	Relationships []Relationship `json:"relationships,omitempty" model:"-"`
}

func (this *Party) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.DisplayName,
			val.NotNilWhen(!forEdit),
			val.When(this.DisplayName != nil,
				val.NotEmpty,
				val.Length(1, 50),
			),
		),
		val.Field(&this.Type,
			val.NotNilWhen(!forEdit),
			val.When(this.Type != nil,
				val.NotEmpty,
				val.OneOf("individual", "company"),
			),
		),

		model.IdPtrValidateRule(&this.OrgId, !forEdit),
		// model.IdPtrValidateRule(&this.Nationality, !forEdit),
		// model.IdPtrValidateRule(&this.Language, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

// type Nationality struct {
// 	Id       string `json:"id"`
// 	Name     string `json:"name"`
// 	ParentId string `json:"parentId"`
// }

// var Nationalitys = []Nationality{
// 	{
// 		Id:       "01HZY8Q3Z0F5X9K4V6G1B2C3D4",
// 		Name:     "Nationality",
// 		ParentId: "",
// 	},
// 	{
// 		Id:       "01HZY8Q50KZAMW4H2KNDS26V3B",
// 		Name:     "Vietnamese",
// 		ParentId: "01HZY8Q3Z0F5X9K4V6G1B2C3D4",
// 	},
// 	{
// 		Id:       "01HZY8Q6FE7SJTCYZTGR0MFTCC",
// 		Name:     "American",
// 		ParentId: "01HZY8Q3Z0F5X9K4V6G1B2C3D4",
// 	},
// 	{
// 		Id:       "01HZY8Q7WXAZFBN7R30FAY1RR7",
// 		Name:     "Japanese",
// 		ParentId: "01HZY8Q3Z0F5X9K4V6G1B2C3D4",
// 	},
// }
