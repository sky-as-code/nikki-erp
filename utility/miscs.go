package utility

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type IEmptyable interface {
	IsEmpty() bool
}

func IsEmpty[T IEmptyable](target *T) bool {
	return target == nil || (*target).IsEmpty()
}

func SetDefaultValue[T any](target *T, defaultValue T) {
	if target == nil {
		target = &defaultValue
	}
}

func SafeVal[T any](source *T, fallbackValue T) T {
	if source != nil {
		return *source
	}
	return fallbackValue
}

func BindIamRequest(ganjingCtx echo.Context, request any) (err error) {
	req := ganjingCtx.Request()
	if req.Method == http.MethodPost && req.ContentLength == 0 {
		// Bind data directly from query params because `ganjingCtx.Bind` skips
		// query params for empty-body POST requests.
		if err = (&echo.DefaultBinder{}).BindQueryParams(ganjingCtx, request); err != nil {
			return err
		}
	} else {
		// Let `ganjingCtx.Bind` bind data.
		if err = ganjingCtx.Bind(request); err != nil {
			return err
		}
	}

	return nil
}
