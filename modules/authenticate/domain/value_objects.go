package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
)

// Account type. Can be Nikki user account or custom account.
// Custom account must be accompanied by a subject source.
// Authenticate Module supports login with accounts from different sources.
type PrincipalType string

const (
	PrincipalTypeNikkiUser = PrincipalType("nikkiuser")
	PrincipalTypeCustom    = PrincipalType("custom")
)

func DefinePrincipalTypeField(fieldName string) *dmodel.FieldBuilder {
	return dmodel.DefineField().Name(fieldName).
		DataType(dmodel.FieldDataTypeEnumString([]string{
			string(PrincipalTypeNikkiUser), string(PrincipalTypeCustom),
		}))
}

func DefinePrincipalDeviceNameField() *dmodel.FieldBuilder {
	return dmodel.DefineField().Name(AttemptFieldDeviceName).
		DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_LONG_NAME_LENGTH))
}

func DefinePrincipalUsernameField(fieldName string) *dmodel.FieldBuilder {
	return dmodel.DefineField().Name(fieldName).
		DataType(dmodel.FieldDataTypeString(5, model.MODEL_RULE_USERNAME_LENGTH))
}

type SendChannel string

const (
	SendChannelEmail = SendChannel("email")
	SendChannelSms   = SendChannel("sms")
)

func DefinePasswordSendChannelField(fieldName string) *dmodel.FieldBuilder {
	return dmodel.DefineField().Name(fieldName).
		DataType(dmodel.FieldDataTypeEnumString([]string{
			string(SendChannelEmail), string(SendChannelSms),
		}))
}
