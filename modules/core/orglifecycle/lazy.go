package orglifecycle

import (
	"sync"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// Lazy builds the service on the first command rather than when the handler subscribes. Handlers
// subscribe during module Init, when a peer module's services may not be registered yet, and
// resolving a resource onion that early would build it ahead of everything it depends on.
func Lazy(build func() (OrgLifecycleService, error)) OrgLifecycleService {
	return &lazyService{build: build}
}

type lazyService struct {
	once  sync.Once
	build func() (OrgLifecycleService, error)
	svc   OrgLifecycleService
	err   error
}

func (this *lazyService) resolve() (OrgLifecycleService, error) {
	// IAM fans one org's lifecycle command out to every module of a wave at once, so builds run
	// concurrently; deps_inject serializes their access to the container.
	this.once.Do(func() {
		this.svc, this.err = this.build()
	})
	return this.svc, this.err
}

func (this *lazyService) InitOrg(ctx corectx.Context, orgId model.Id) error {
	svc, err := this.resolve()
	if err != nil {
		return err
	}
	return svc.InitOrg(ctx, orgId)
}

func (this *lazyService) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	svc, err := this.resolve()
	if err != nil {
		return err
	}
	return svc.CleanUpOrg(ctx, orgId)
}
