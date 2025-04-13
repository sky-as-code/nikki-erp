package utility

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func BodyDump(ignore ...string) echo.MiddlewareFunc {

	ignoreMap := map[string]int{}
	for _, key := range ignore {
		ignoreMap[key] = 1
	}

	return middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		if c.Request().Method == "POST" {
			if _, found := ignoreMap[c.Path()]; !found {
				log.Printf("%v %v\nIn:  %v\nOut: %v\n", c.Request().Method, c.Path(), string(reqBody), string(resBody))
			}
		}
	})
}
