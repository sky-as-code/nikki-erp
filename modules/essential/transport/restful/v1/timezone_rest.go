package v1

import (
	"github.com/labstack/echo/v5"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/timezone"
)

func NewTimezoneRest(timezoneSvc it.TimezoneAppService) *TimezoneRest {
	return &TimezoneRest{timezoneSvc: timezoneSvc}
}

// TimezoneRest serves the timezone catalogue (?keyword=&language_code=&page=&size=).
type TimezoneRest struct {
	timezoneSvc it.TimezoneAppService
}

func (this *TimezoneRest) Search(echoCtx *echo.Context) error {
	payload := map[string]any{}
	for name, values := range echoCtx.QueryParams() {
		if len(values) > 0 {
			payload[name] = values[0]
		}
	}
	return dynamicresource.ServeAction(echoCtx, "search timezones", payload,
		this.timezoneSvc.SearchTimezonesAction, dynamicresource.Identity)
}
