package event

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type PricingChangedType string

const (
	PricingChangedPricelist      PricingChangedType = "sales_pricelist"
	PricingChangedPricelistItem  PricingChangedType = "sales_pricelist_item"
	PricingChangedCombo          PricingChangedType = "sales_combo"
	PricingChangedComboComponent PricingChangedType = "sales_combo_component"
)

type PricingChangedEvent struct {
	Type  PricingChangedType `json:"type"`
	OrgId string             `json:"org_id"`
	Ids   []string           `json:"ids,omitempty"`
}

type PricingChangedEventPublisher interface {
	Publish(ctx corectx.Context, eventType PricingChangedType, orgId string, ids ...string) error
	PublishAsync(ctx corectx.Context, eventType PricingChangedType, orgId string, ids ...string)
}
