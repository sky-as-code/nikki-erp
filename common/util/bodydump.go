package util

import (
	"log"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func BodyDump(ignore ...string) echo.MiddlewareFunc {

	ignoreMap := map[string]int{}
	for _, key := range ignore {
		ignoreMap[key] = 1
	}

	return middleware.BodyDump(func(c *echo.Context, reqBody, resBody []byte, err error) {
		if (*c).Request().Method == "POST" {
			if _, found := ignoreMap[(*c).Path()]; !found {
				log.Printf("%v %v\nIn:  %v\nOut: %v\nErr: %v\n", (*c).Request().Method, (*c).Path(), string(reqBody), string(resBody), err)
			}
		}
	})
}
