package role_suite

import (
	"github.com/sky-as-code/nikki-erp/common/model"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateRoleSuiteCommand) ToRoleSuite() *domain.RoleSuite {
	roleSuite := &domain.RoleSuite{}
	model.MustCopy(this, roleSuite)
	return roleSuite
}

func (this UpdateRoleSuiteCommand) ToRoleSuite() *domain.RoleSuite {
	roleSuite := &domain.RoleSuite{}
	model.MustCopy(this, roleSuite)
	return roleSuite
}
