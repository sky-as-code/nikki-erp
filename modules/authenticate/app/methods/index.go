package methods

import (
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
)

var allMethods = []it.LoginMethod{
	// 1. mTLS
	// 2. QR Code
	// 3. Hard key / Passkey
	// 4. Password + Temp Password
	&LoginMethodPassword{},
	// 5. Captcha
	&LoginMethodCaptcha{},
	// 6. Password OTP
	&LoginMethodOtpCode{},
}
var methodMap map[string]it.LoginMethod
var methodNames []string

func init() {
	methodMap = make(map[string]it.LoginMethod, len(allMethods))
	for _, m := range allMethods {
		methodMap[m.Name()] = m
		methodNames = append(methodNames, m.Name())
	}
}

func AllLoginMethods() []it.LoginMethod {
	return allMethods
}

func GetLoginMethod(name string) it.LoginMethod {
	return methodMap[name]
}
