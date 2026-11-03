package dynamicresource

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// customActionName is a custom route's action for its span name: the literal path segments,
// never the ":param" ones, so no id reaches the name. The resource base is named by its type.
func customActionName(def RouteDefinition) string {
	literal := []string{}
	for _, segment := range strings.Split(def.Path, "/") {
		if segment != "" && !strings.HasPrefix(segment, ":") {
			literal = append(literal, segment)
		}
	}
	if len(literal) == 0 {
		return strings.ToLower(string(def.ActionType))
	}
	return strings.Join(literal, ".")
}

// traceCustomAction gives a custom action the same {schema}.{action} span the built-in ones open
// in the application default. A custom action is a method on a derived service, so its one
// generic boundary is the route; a service-only custom method records under its caller's span.
func traceCustomAction(schemaName string, action string, handler echo.HandlerFunc) echo.HandlerFunc {
	return func(echoCtx *echo.Context) error {
		reqCtx, isReqCtx := echoCtx.Request().Context().(corectx.Context)
		if !isReqCtx {
			return handler(echoCtx)
		}
		end := StartActionSpan(reqCtx, schemaName, action)
		err := handler(echoCtx)
		_, status := echo.ResolveResponseStatus(echoCtx.Response(), err)
		switch {
		case status < http.StatusInternalServerError:
			end(nil)
		case err != nil:
			end(err)
		default:
			end(errors.Errorf("%s.%s answered %d", schemaName, action, status))
		}
		return err
	}
}
