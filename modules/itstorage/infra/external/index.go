// Package external binds IT Storage's local ports to the services other modules publish.
//
// This is the ONLY package in IT Storage that may import another module.
package external

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itExt "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/external"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

func InitExternalServices() error {
	return deps.Register(func(quotaSvc itSettings.QuotaAppService) itExt.QuotaExtService {
		return quotaSvc
	})
}
