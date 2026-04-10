package authtoken

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitSubModule() error {
	err := deps.Register(NewAuthTokenServiceImpl)
	return err
}
