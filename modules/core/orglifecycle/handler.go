package orglifecycle

import (
	"context"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

// RegisterHandlers subscribes moduleName's `initOrg` and `cleanUpOrg` handlers. Called from the
// module's transport/cqrs Init.
func RegisterHandlers(ctx context.Context, bus cqrs.CqrsBus, moduleName string, svc OrgLifecycleService) error {
	if moduleName == "" {
		return errors.New("orglifecycle: a handler needs the name of the module it answers for")
	}
	if svc == nil {
		return errors.Errorf("orglifecycle: module %s has no org lifecycle service", moduleName)
	}

	handler := &lifecycleHandler{moduleName: moduleName, svc: svc}
	return bus.SubscribeRequests(ctx,
		newModuleHandler(moduleName, ActionInitOrg, handler.handle),
		newModuleHandler(moduleName, ActionCleanUpOrg, handler.handle),
	)
}

// newModuleHandler builds a handler subscribed under this module's request type. See
// usagecheck.newModuleHandler: cqrs.NewHandler derives the topic from a zero-valued request,
// which names no module.
func newModuleHandler(
	moduleName string, action string,
	handle func(context.Context, *cqrs.RequestPacket[OrgLifecycleCommand]) (*cqrs.Reply[OrgLifecycleResult], error),
) cqrs.RequestHandler {
	return &moduleScopedHandler{
		RequestHandler: cqrs.NewHandler(handle),
		moduleName:     moduleName,
		action:         action,
	}
}

type moduleScopedHandler struct {
	cqrs.RequestHandler

	moduleName string
	action     string
}

func (this *moduleScopedHandler) NewRequest() any {
	return &OrgLifecycleCommand{TargetModule: this.moduleName, Action: this.action}
}

type lifecycleHandler struct {
	moduleName string
	svc        OrgLifecycleService
}

func (this *lifecycleHandler) handle(
	ctx context.Context, packet *cqrs.RequestPacket[OrgLifecycleCommand],
) (*cqrs.Reply[OrgLifecycleResult], error) {
	return this.answer(ctx, packet.Request())
}

// answer is separate from handle so it can be tested without a bus packet.
func (this *lifecycleHandler) answer(
	ctx context.Context, cmd *OrgLifecycleCommand,
) (*cqrs.Reply[OrgLifecycleResult], error) {
	if err := this.validate(cmd); err != nil {
		return nil, err
	}

	orgId := model.Id(cmd.OrgId)
	svcCtx := ServiceContextOf(ctx, this.moduleName, orgId, model.Id(cmd.RequestedBy))

	var err error
	if cmd.Action == ActionInitOrg {
		err = this.svc.InitOrg(svcCtx, orgId)
	} else {
		err = this.svc.CleanUpOrg(svcCtx, orgId)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "orglifecycle: module %s %s org %s", this.moduleName, cmd.Action, cmd.OrgId)
	}

	return &cqrs.Reply[OrgLifecycleResult]{
		Result: OrgLifecycleResult{RequestId: cmd.RequestId, Module: this.moduleName},
	}, nil
}

func (this *lifecycleHandler) validate(cmd *OrgLifecycleCommand) error {
	if cmd.TargetModule != this.moduleName {
		return errors.Errorf("orglifecycle: module %s received a command addressed to %s",
			this.moduleName, cmd.TargetModule)
	}
	if cmd.Action != ActionInitOrg && cmd.Action != ActionCleanUpOrg {
		return errors.Errorf("orglifecycle: module %s received unknown action %q", this.moduleName, cmd.Action)
	}
	if cmd.OrgId == "" {
		return errors.Errorf("orglifecycle: module %s received a command naming no org", this.moduleName)
	}
	return nil
}
