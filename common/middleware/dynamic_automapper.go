package middleware

import (
	"encoding/json"
	"reflect"

	"github.com/labstack/echo/v4"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicentity/model"
)

type creatorFn[T any] func() T
type RestHandlerFn = func(request any, echoCtx echo.Context) error

func DynamicAutoMapper[T any](handlerFn RestHandlerFn) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		requestFields, found := echoCtx.Get("nikkiDynamicRequestFields").(dmodel.EntityMap)
		if !found {
			var err error
			requestFields, err = ExtractRequestFields(echoCtx)
			if err != nil {
				return err
			}
		}

		var dest *T = dmodel.EntityMapToStruct[T](requestFields)

		return handlerFn(dest, echoCtx)
	}
}

func DynamicAutoMapper2[T any](create creatorFn[T], handlerFn RestHandlerFn) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		requestFields, found := echoCtx.Get("nikkiDynamicRequestFields").(dmodel.EntityMap)
		if !found {
			var err error
			requestFields, err = ExtractRequestFields(echoCtx)
			if err != nil {
				return err
			}
		}

		var dest any = create()
		destValue := reflect.ValueOf(dest)
		isPointer := destValue.Kind() == reflect.Ptr

		// Ensure we have a pointer for JSON unmarshaling
		var destPtr any
		if isPointer {
			destPtr = dest
		} else {
			ptrValue := reflect.New(destValue.Type())
			ptrValue.Elem().Set(destValue)
			destPtr = ptrValue.Interface()
		}

		// Map requestFields to dest using JSON tags
		// Marshal map to JSON, then unmarshal into struct
		jsonBytes, err := json.Marshal(requestFields)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(jsonBytes, destPtr); err != nil {
			return err
		}

		// Extract value if original was not a pointer
		if !isPointer {
			dest = reflect.ValueOf(destPtr).Elem().Interface()
		} else {
			dest = destPtr
		}

		return handlerFn(dest, echoCtx)
	}
}
