package slabreach

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type SlaBreachService interface {
	CreateSlaBreach(ctx corectx.Context, cmd CreateSlaBreachCommand) (*CreateSlaBreachResult, error)
	DeleteSlaBreach(ctx corectx.Context, cmd DeleteSlaBreachCommand) (*DeleteSlaBreachResult, error)
	GetSlaBreach(ctx corectx.Context, query GetSlaBreachQuery) (*GetSlaBreachResult, error)
	SlaBreachExists(ctx corectx.Context, query SlaBreachExistsQuery) (*SlaBreachExistsResult, error)
	SearchSlaBreaches(ctx corectx.Context, query SearchSlaBreachesQuery) (*SearchSlaBreachesResult, error)
	UpdateSlaBreach(ctx corectx.Context, cmd UpdateSlaBreachCommand) (*UpdateSlaBreachResult, error)
}
