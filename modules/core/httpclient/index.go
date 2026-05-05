package httpclient

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/client"
)

func InitSubModule() error {
	err := deps.Register(client.NewCoreHttpClient)
	return err
}
