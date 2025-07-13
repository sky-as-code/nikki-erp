package tag

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
)

func EnumToTag(src *enum.Enum) *Tag {
	if src == nil {
		return nil
	}
	return &Tag{
		ModelBase: src.ModelBase,
		Label:     src.Label,
		Type:      src.Type,
	}
}

func EnumsToTags(srcs []enum.Enum) []Tag {
	if srcs == nil {
		return nil
	}
	return array.Map(srcs, func(src enum.Enum) Tag {
		return *EnumToTag(&src)
	})
}
