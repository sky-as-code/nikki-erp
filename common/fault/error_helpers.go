package fault

import (
	goerr "errors"
	"fmt"
	"reflect"

	"go.bryk.io/pkg/errors"
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

func RecoverPanicFailedTo(err any, action string) error {
	return RecoverPanic(err, fmt.Sprintf("failed to %s", action))
}

func RecoverPanicf(err any, errMsg string, args ...any) error {
	return RecoverPanic(err, fmt.Sprintf(errMsg, args...))
}

func RecoverPanic(err any, errMsg string) error {
	if err != nil {
		originErr, isOk := err.(error)
		if isOk {
			return errors.Wrap(originErr, errMsg)
		} else {
			originErr = fmt.Errorf("%v", err)
			return errors.Wrap(originErr, errMsg)
		}
	}
	return nil
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
