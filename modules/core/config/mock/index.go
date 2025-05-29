package mock

import (
	"github.com/golang/mock/gomock"

	. "github.com/sky-as-code/nikki-erp/modules/core/config"
)

func NewConfigSvcMock(ctrl *gomock.Controller) (svc ConfigService, loader *MockConfigLoader) {
	loader = NewMockConfigLoader(ctrl)
	svc = NewConfigService(loader)
	return
}
