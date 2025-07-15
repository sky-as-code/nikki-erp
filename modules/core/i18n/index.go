package i18n

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/modules/core/i18n/app"
	"github.com/sky-as-code/nikki-erp/modules/core/i18n/transport"
)

func InitSubModule() error {
	err := stdErr.Join(
		app.InitServices(),
		transport.InitTransport(),
	)

	return err
}
