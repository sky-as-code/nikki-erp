package orglifecycle

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

// InitSubModule registers the one dispatcher IAM resolves to run org lifecycle actions. Each
// module subscribes its own handlers with RegisterHandlers.
func InitSubModule() error {
	return deps.Register(NewDispatcher)
}

// RegisterNoop subscribes a module that holds no org-scoped data. It still answers, so iam can
// tell "nothing to do" from "never subscribed".
func RegisterNoop(moduleName string) error {
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		return RegisterHandlers(context.Background(), bus, moduleName, NoopOrgLifecycleService{})
	})
}
