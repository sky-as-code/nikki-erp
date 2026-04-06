package login

import (
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
)

func (this CreateLoginAttemptCommand) ToLoginAttempt() *domain.LoginAttempt {
	attempt := domain.NewLoginAttempt()
	attempt.SetDeviceIp(this.DeviceIp)
	attempt.SetDeviceName(this.DeviceName)
	attempt.SetDeviceLocation(this.DeviceLocation)
	attempt.SetSubjectType(&this.SubjectType)
	attempt.SetSubjectSourceRef(this.SubjectSourceRef)
	attempt.SetUsername(&this.Username)
	return attempt
}

func (this UpdateLoginAttemptCommand) ToLoginAttempt() *domain.LoginAttempt {
	attempt := domain.NewLoginAttempt()
	attempt.SetId(&this.Id)
	attempt.SetIsGenuine(this.IsGenuine)
	attempt.SetCurrentMethod(this.CurrentMethod)
	attempt.SetStatus(this.Status)
	return attempt
}
