package job

import (
	"context"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// Scoping a background sweep to the data it may touch.
//
// # Why this seam exists
//
// A cron job has no request behind it, so its context carries no user, no organization and — in a
// multi-tenant build — no tenant. In nikkierp that is fine: a sweep runs once and sees everything.
// In coremart it is fatal. Every tenant-scoped table refuses a read whose context names no tenant,
// and the repository PANICS rather than returning an error, because a query without a tenant would
// silently read across all of them.
//
// That is why the cron scheduler sat unstarted for so long: starting it made every sweep in the
// codebase panic on its first tick, including paymentinvoice's three.
//
// # The seam
//
// A sweep is written once, against a context it is handed, and never mentions tenancy. SweepScoper
// decides what "once" means: in nikkierp it runs the sweep a single time on a plain background
// context; in coremart it runs the sweep once per enabled tenant, each with a context scoped to it.
//
// The default below is the nikkierp behaviour. coremart replaces it during Init with a fanout, so
// the sweeps themselves — in sales, in paymentinvoice — need no build tags and no knowledge of which
// deployment they are in.

// TenantlessSweep is a background sweep, written as though tenancy did not exist.
type TenantlessSweep func(ctx corectx.Context) error

// SweepScoper turns a sweep into the handler the scheduler runs.
//
// It exists so that "run this sweep" can mean different things in different builds without the
// sweep knowing. See the package comment above.
type SweepScoper func(sweep TenantlessSweep, jobName string) JobHandleFn

// sweepScoper is the active strategy. Package-level rather than injected, because a sweep is
// registered from a module's OnAppStarted and the scoping decision belongs to the deployment, not
// to any one module that happens to own a sweep.
var sweepScoper SweepScoper = singleScopeSweep

// SetSweepScoper installs a deployment's scoping strategy.
//
// Called once, from Init, BEFORE any module registers a job — a scoper installed later would leave
// the jobs registered before it running unscoped, which in coremart means panicking on every tick.
func SetSweepScoper(scoper SweepScoper) {
	if scoper != nil {
		sweepScoper = scoper
	}
}

// ScopeSweep wraps a sweep for the scheduler, using whatever strategy this deployment installed.
//
// Every module registering a cron sweep goes through this rather than building a JobHandleFn
// directly, which is what makes the tenant fanout apply everywhere instead of only where somebody
// remembered it.
func ScopeSweep(sweep TenantlessSweep, jobName string) JobHandleFn {
	return sweepScoper(sweep, jobName)
}

// singleScopeSweep runs a sweep exactly once per tick, on a plain context.
//
// The nikkierp default, and the correct behaviour for a build with no tenants: there is one dataset,
// so one run covers it.
func singleScopeSweep(sweep TenantlessSweep, _ string) JobHandleFn {
	return func(ctx context.Context, _ *string) error {
		return sweep(corectx.NewRequestContext(ctx))
	}
}
