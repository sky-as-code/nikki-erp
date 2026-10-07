package event

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type CatalogChangedType string

const (
	CatalogChangedCategory        CatalogChangedType = "product_category"
	CatalogChangedProductTemplate CatalogChangedType = "product_template"
	CatalogChangedProductVariant  CatalogChangedType = "product_variant"
)

type CatalogChangedEvent struct {
	Type  CatalogChangedType `json:"type"`
	OrgId string             `json:"org_id"`
	Ids   []string           `json:"ids,omitempty"`
}

type CatalogChangedEventPublisher interface {
	Publish(ctx corectx.Context, eventType CatalogChangedType, orgId string, ids ...string) error
	PublishAsync(ctx corectx.Context, eventType CatalogChangedType, orgId string, ids ...string)
}
