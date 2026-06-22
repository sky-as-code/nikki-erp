package language

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/modules/core/language/app"
	"github.com/sky-as-code/nikki-erp/modules/core/language/transport"
)

func InitSubModule() error {
	err := stdErr.Join(
		app.InitServices(),
		transport.InitTransport(),
	)

	return err
}
