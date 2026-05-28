package enum

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type EnumService interface {
	CreateEnum(ctx corectx.Context, cmd CreateEnumCommand) (*CreateEnumResult, error)
	UpdateEnum(ctx corectx.Context, cmd UpdateEnumCommand) (*UpdateEnumResult, error)
	DeleteEnum(ctx corectx.Context, cmd DeleteEnumCommand) (*DeleteEnumResult, error)
	GetEnum(ctx corectx.Context, query GetEnumQuery) (*GetEnumResult, error)
	SearchEnums(ctx corectx.Context, query SearchEnumsQuery) (*SearchEnumsResult, error)
	EnumExists(ctx corectx.Context, query EnumExistsQuery) (*EnumExistsResult, error)
}
