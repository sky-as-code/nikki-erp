package fault

import (
	goerr "errors"
	"fmt"
	"reflect"
)

func PanicOnErr(err error) {
	if !isNil(err) {
		panic(err)
	}
}

func isNil(input interface{}) bool {
	if input == nil {
		return true
	}
	kind := reflect.ValueOf(input).Kind()
	switch kind {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan:
		return reflect.ValueOf(input).IsNil()
	default:
		return false
	}
}

func JoinErrors(errArr []error) error {
	if len(errArr) == 0 {
		return nil
	}
	errMsg := errArr[0].Error()
	for i := 1; i < len(errArr); i++ {
		errMsg = fmt.Sprintf("%s; %s", errMsg, errArr[i].Error())
	}
	return goerr.New(errMsg)
}
