package enum

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/copier"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/ent"
)

func (this CreateEnumCommand) ToEnum() *Enum {
	return &Enum{
		Label: &this.Label,
		Type:  &this.EnumType,
		Value: &this.Value,
	}
}

func (this UpdateEnumCommand) ToEnum() *Enum {
	return &Enum{
		ModelBase: model.ModelBase{
			Id:   &this.Id,
			Etag: &this.Etag,
		},
		Label: this.Label,
		Value: this.Value,
	}
}

func EntToEnum(dbEnum *ent.Enum) *Enum {
	return &Enum{
		ModelBase: model.ModelBase{
			Id:   &dbEnum.ID,
			Etag: &dbEnum.Etag,
		},
		Label: &dbEnum.Label,
		Value: &dbEnum.Value,
		Type:  &dbEnum.Type,
	}
}

func EntToEnums(dbEnums []*ent.Enum) []Enum {
	if dbEnums == nil {
		return nil
	}
	return array.Map(dbEnums, func(entEnum *ent.Enum) Enum {
		return *EntToEnum(entEnum)
	})
}

func AnyToEnum(dbEnum any) *Enum {
	domainEnum := Enum{}
	err := copier.Copy(dbEnum, &domainEnum)
	fault.PanicOnErr(err)
	return &domainEnum
}

func AnyToEnums(dbEnums []any) []Enum {
	if dbEnums == nil {
		return nil
	}
	return array.Map(dbEnums, func(dbEnum any) Enum {
		return *AnyToEnum(dbEnum)
	})
}
