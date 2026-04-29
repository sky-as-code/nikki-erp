package external

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itExt "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/external"
	itSet "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/userpref"
)

func InitExternalServices() error {
	return stdErr.Join(
		deps.Register(func(userPrefSvc itSet.UserPreferenceUiDomainService) itExt.UserPreferenceUiDomainService {
			return userPrefSvc
		}),
	)
}
