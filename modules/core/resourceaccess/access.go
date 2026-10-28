// Package resourceaccess is the core port through which a module asks whether the caller may read
// a record of another module's resource, without importing that module. The dynamic resource
// engine implements it from its engine registry (`dynengine_{schema}`), so any resource built on
// the engine can be asked about.
package resourceaccess

import (
	"sync"

	"go.bryk.io/pkg/errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type ReadAsCallerQuery struct {
	// ResourceCode is the schema name, which is also the permission resource code.
	ResourceCode string
	ResourceId   model.Id
	// OrgId is the org the caller names; ignored by a resource that is not org-scoped.
	OrgId *model.Id
}

type ReadAsCallerResult struct {
	// Exists is false for a record that is gone. The caller then held the resource-level read
	// grant in OrgId, which is all a deleted record can be checked against.
	Exists bool
	// MaskedFields lists the fields of the record the caller may not read.
	MaskedFields []string
}

// ResourceAccess runs the target resource's own read authorization as the caller: the same
// permission, org membership, org confinement and field masking a GET of that record applies.
// A refusal comes back as client errors; a Go error is a server fault.
type ResourceAccess interface {
	ReadAsCaller(ctx corectx.Context, query ReadAsCallerQuery) (*ReadAsCallerResult, ft.ClientErrors, error)

	// LabelsAsCaller answers the display label (the schema's record_label_field, in the caller's
	// language) of each record in one batched read, run as the caller. An id the caller may not
	// read, or that no longer exists, is left out; a resource with no record label answers none.
	LabelsAsCaller(ctx corectx.Context, query LabelsQuery) (map[string]string, error)
}

type LabelsQuery struct {
	ResourceCode string
	Ids          []model.Id
	// OrgId is the org the caller names; ignored by a resource that is not org-scoped.
	OrgId *model.Id
}

// LazyResourceAccess resolves the port on first use, so a module built before the engine's Init
// still reaches it. Unlike a recording port it fails closed: with no implementation registered,
// every check is an error rather than an approval.
type LazyResourceAccess struct {
	once   sync.Once
	access ResourceAccess
}

func NewLazyResourceAccess() *LazyResourceAccess {
	return &LazyResourceAccess{}
}

func (this *LazyResourceAccess) ReadAsCaller(
	ctx corectx.Context, query ReadAsCallerQuery,
) (*ReadAsCallerResult, ft.ClientErrors, error) {
	if this.resolve() == nil {
		return nil, nil, errors.New("no resourceaccess.ResourceAccess is registered")
	}
	return this.access.ReadAsCaller(ctx, query)
}

// LabelsAsCaller answers no labels without an implementation: a label is a convenience, and the
// caller falls back to the raw id.
func (this *LazyResourceAccess) LabelsAsCaller(ctx corectx.Context, query LabelsQuery) (map[string]string, error) {
	if this.resolve() == nil {
		return map[string]string{}, nil
	}
	return this.access.LabelsAsCaller(ctx, query)
}

func (this *LazyResourceAccess) resolve() ResourceAccess {
	this.once.Do(func() {
		_ = deps.Invoke(func(access ResourceAccess) {
			this.access = access
		})
	})
	return this.access
}
