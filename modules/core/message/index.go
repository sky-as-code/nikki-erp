package message

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/message/transports"
)

func InitSubModule() error {
	return deps.Register(
		transports.NewMqttTransport,
		transports.NewRedisTransport,
		transports.NewGoChannleTransport,
	)
}
