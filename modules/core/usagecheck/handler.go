package usagecheck

import (
	"context"

	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

// SubscribeHandler subscribes moduleName's answer to usage checks. Called from the dependant
// module's Init, once per module however many resources it checks: the registry dispatches
// within the module, so one subscription serves every upstream resource it answers about.
func SubscribeHandler(
	ctx context.Context, bus cqrs.CqrsBus, moduleName string, registry *CheckerRegistry,
) error {
	if moduleName == "" {
		return errors.New("usagecheck: a usage handler needs the name of the module it answers for")
	}
	if registry == nil {
		return errors.Errorf("usagecheck: module %s has no checker registry", moduleName)
	}

	handler := &checkUsageHandler{moduleName: moduleName, registry: registry}
	return bus.SubscribeRequests(ctx, newModuleHandler(moduleName, handler.handle))
}

// newModuleHandler builds a handler subscribed under THIS module's request type.
//
// cqrs.NewHandler derives the topic from a zero-valued request, so it cannot be used here: the
// command's type is a function of its TargetModule, and a zero value names no module. Every
// module would subscribe to the same "_usage.checkResourceUsage" topic — the second to start
// would collide with the first, and no lookup for "sales_usage.checkResourceUsage" would ever
// find a handler.
func newModuleHandler(
	moduleName string,
	handle func(context.Context, *cqrs.RequestPacket[CheckResourceUsageCommand]) (
		*cqrs.Reply[CheckResourceUsageResult], error),
) cqrs.RequestHandler {
	return &moduleScopedHandler{
		moduleName:  moduleName,
		RequestHandler: cqrs.NewHandler(handle),
	}
}

// moduleScopedHandler answers NewRequest with a command already addressed to its own module, so
// the bus reads this module's topic from it. Everything else is the generic handler's.
type moduleScopedHandler struct {
	cqrs.RequestHandler

	moduleName string
}

func (this *moduleScopedHandler) NewRequest() any {
	return &CheckResourceUsageCommand{TargetModule: this.moduleName}
}

type checkUsageHandler struct {
	moduleName string
	registry   *CheckerRegistry
}

// handle answers one command, one entry per requested resource.
//
// Every failure returns an error rather than a negative answer: the caller treats an unanswered
// check as blocking, so an error refuses the delete, while a false would permit it. That is the
// asymmetry the whole design turns on — refusing a delete that would have been fine is
// recoverable, permitting one that breaks a reference is not.
func (this *checkUsageHandler) handle(
	ctx context.Context, packet *cqrs.RequestPacket[CheckResourceUsageCommand],
) (*cqrs.Reply[CheckResourceUsageResult], error) {
	return this.answer(ctx, packet.Request())
}

// answer holds everything the handler does with a command. It is separate from handle so the
// behaviour can be tested without building a bus packet, whose fields are unexported.
func (this *checkUsageHandler) answer(
	ctx context.Context, cmd *CheckResourceUsageCommand,
) (*cqrs.Reply[CheckResourceUsageResult], error) {
	if err := this.validate(cmd); err != nil {
		return nil, err
	}

	reqCtx, ok := ctx.(corectx.Context)
	if !ok {
		return nil, errors.Errorf(
			"usagecheck: module %s received a usage check without a request context", this.moduleName)
	}

	results := make([]ResourceUsageItem, 0, len(cmd.Resources))
	for _, ref := range cmd.Resources {
		checker, found := this.registry.Checker(cmd.SourceModule, ref.ResourceName)
		if !found {
			// The owning module believes this module consults it about this resource and it does
			// not. Silence here would read as "not used" and permit the delete, so it is an error:
			// either the dependant list names a module that should not be on it, or this module
			// forgot to register a checker.
			return nil, errors.Errorf(
				"usagecheck: module %s has no checker for %s.%s, asked by %s",
				this.moduleName, cmd.SourceModule, ref.ResourceName, cmd.SourceModule)
		}

		used, usedBy, err := checker.IsUsed(reqCtx, ref)
		if err != nil {
			return nil, errors.Wrapf(err, "usagecheck: module %s checking %s.%s",
				this.moduleName, cmd.SourceModule, ref.ResourceName)
		}

		results = append(results, ResourceUsageItem{
			ResourceName: ref.ResourceName,
			Identifier:   ref.Identifier,
			IsUsed:       used,
			UsedBy:       usedBy,
		})
	}

	return &cqrs.Reply[CheckResourceUsageResult]{
		Result: CheckResourceUsageResult{
			RequestId: cmd.RequestId,
			Module:    this.moduleName,
			Results:   results,
		},
	}, nil
}

func (this *checkUsageHandler) validate(cmd *CheckResourceUsageCommand) error {
	if cmd.SourceModule == "" {
		return errors.Errorf("usagecheck: module %s received a usage check naming no source module",
			this.moduleName)
	}
	if cmd.TargetModule != this.moduleName {
		// Addressed elsewhere. Answering anyway would report this module's usage under another
		// module's name, and the caller would take it as that module's answer.
		return errors.Errorf("usagecheck: module %s received a usage check addressed to %s",
			this.moduleName, cmd.TargetModule)
	}
	if len(cmd.Resources) == 0 {
		return errors.Errorf("usagecheck: module %s received a usage check listing no resources",
			this.moduleName)
	}
	for _, ref := range cmd.Resources {
		if ref.ResourceName == "" {
			return errors.Errorf("usagecheck: module %s received a resource with no name",
				this.moduleName)
		}
		if len(ref.Identifier) == 0 {
			return errors.Errorf("usagecheck: module %s received resource %s with no identifier",
				this.moduleName, ref.ResourceName)
		}
	}
	return nil
}
