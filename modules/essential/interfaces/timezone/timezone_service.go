// Package timezone declares the search over the hardcoded timezone catalogue.
package timezone

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	tz "github.com/sky-as-code/nikki-erp/modules/essential/domain/timezone"
)

type SearchTimezonesQuery struct {
	Keyword      string `json:"keyword"`
	LanguageCode string `json:"language_code"`
	Page         int    `json:"page"`
	Size         int    `json:"size"`
}

type SearchTimezonesResultData struct {
	Items []tz.Zone `json:"items"`
	Total int       `json:"total"`
}

type SearchTimezonesResult = dyn.OpResult[SearchTimezonesResultData]

// TimezoneAppService serves the catalogue to any signed-in caller: it is reference data, like the
// language bundles, so it asserts no permission.
type TimezoneAppService interface {
	SearchTimezones(ctx corectx.Context, query SearchTimezonesQuery) (*SearchTimezonesResult, error)
	SearchTimezonesAction(ctx corectx.Context, params dmodel.DynamicFields) (*SearchTimezonesResult, error)
}
