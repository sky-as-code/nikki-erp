package context

import (
	"context"

	"github.com/labstack/echo/v5"
	"go.bryk.io/pkg/errors"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type Context interface {
	context.Context

	InnerContext() context.Context
	GetLogger() logging.LoggerService
	SetLogger(logger logging.LoggerService)
	GetDbTranx() db.DbTransaction
	SetDbTranx(trx db.DbTransaction)
	GetDomainConstraints() dmodel.DynamicFields
	SetDomainConstraints(constraints dmodel.DynamicFields)
	GetModuleName() string
	GetPermissions() ContextPermissions
	SetPermissions(permissions ContextPermissions)
	GetUser() dmodel.DynamicFields
	SetUser(user dmodel.DynamicFields)
	// Replace current inner context with a new one that has the given key and value.
	WithValue(key, val any)
}

// PrincipalKind says what sort of actor is executing, so that a job or a consumer is
// distinguishable from an unauthenticated caller rather than looking identical to one.
//
// This is not IAM's PrincipalType (`nikkiuser` | `custom`), which describes where a login
// credential came from - a different axis, and in a module core cannot import.
type PrincipalKind string

const (
	// PrincipalKindUser is a person acting through a client.
	PrincipalKindUser = PrincipalKind("user")
	// PrincipalKindService is a workload acting on the system's behalf: a scheduled job, a
	// message consumer, a device. It authorizes exactly like a user, on entitlements it was
	// granted - never by virtue of being internal.
	PrincipalKindService = PrincipalKind("service")
	// PrincipalKindSystem is reserved for future platform-level operations. It deliberately
	// carries NO bypass behaviour; treat it as unauthorized until that is designed.
	PrincipalKindSystem = PrincipalKind("system")
)

// Principal is the authenticated actor behind the current execution.
//
// A zero Principal means "nobody was authenticated". Authorization must fail closed on it rather
// than reading absence as permission - see requestguard.AssertPermission.
type Principal struct {
	Kind PrincipalKind `json:"kind"`
	Id   model.Id      `json:"id"`
	// OrgId is the single org a service principal acts within. A service holds no org
	// membership, so this is what org-scoped work is checked against; nil for a user, whose
	// reach comes from UserOrgIds instead.
	OrgId *model.Id `json:"org_id"`
	// DisplayName is for audit and logs only. Never make a security decision from it.
	DisplayName string `json:"display_name"`
}

// IsZero reports whether no principal was established.
func (this Principal) IsZero() bool {
	return this.Kind == "" && this.Id == ""
}

type ContextPermissions struct {
	IsOwner      bool
	Entitlements ds.Set[string]
	UserId       model.Id `json:"user_id"`
	// The actor executing this request: a user, or a service acting for the system.
	Principal Principal `json:"principal"`
	// The orgs that user belongs to (if any)
	UserOrgIds ds.Set[model.Id] `json:"user_org_ids"`
	// The org unit that user belongs to (if any)
	OrgUnitId *model.Id `json:"org_unit_id"`
	// The org that the org unit belongs to (if user belongs to an org unit)
	OrgUnitOrgId *model.Id `json:"org_unit_org_id"`
}

func NewRequestContext(ctx context.Context) Context {
	return &RequestContext{
		Context: ctx,
	}
}

func NewRequestContextM(ctx context.Context, moduleName string) Context {
	return &RequestContext{
		Context:    ctx,
		moduleName: moduleName,
	}
}

// NewRequestContextF builds an empty context carrying domain constraints.
//
// The constraints go through SetDomainConstraints rather than a struct field: GetDomainConstraints
// reads them back off the inner context's values, so a field set here would never be read.
func NewRequestContextF(ctx context.Context, moduleName string, domainConstraints dmodel.DynamicFields) Context {
	out := &RequestContext{
		Context:    ctx,
		moduleName: moduleName,
	}
	out.SetDomainConstraints(domainConstraints)
	return out
}

// CloneRequestContext returns a copy of ctx that can be given a different transaction without
// disturbing the original. Use it to derive a scoped context from a live request; the New*
// constructors above build an empty one, for code with no caller behind it.
//
// It copies every field the copy might be used with, the caller's identity above all: basemodel
// stamps created_by/updated_by from GetPermissions().UserId, so a clone that dropped permissions
// would write those columns null and report nothing. Domain constraints live in the inner
// context's values, so copying Context carries them along.
//
// A field added to RequestContext must be added here too.
func CloneRequestContext(ctx Context) Context {
	return &RequestContext{
		Context:     ctx.InnerContext(),
		logger:      ctx.GetLogger(),
		repoTrx:     ctx.GetDbTranx(),
		moduleName:  ctx.GetModuleName(),
		permissions: ctx.GetPermissions(),
		user:        ctx.GetUser(),
	}
}

// Returns pointer to an instance of RequestContext if it exists, otherwise returns an error.
func AsRequestContext(echoCtx *echo.Context) (Context, error) {
	reqCtx, isReqCtx := echoCtx.Request().Context().(Context)
	if !isReqCtx {
		return nil, errors.New("Must have RequestContextMiddleware2 before calling this function")
	}
	return reqCtx, nil
}

type contextKey struct {
	name string
}

var CtxKeyDomainConstraints = contextKey{"domain_constraints"}

type RequestContext struct {
	context.Context

	logger logging.LoggerService

	// The transaction object that Repository Layer can use to perform atomic database operations.
	repoTrx     db.DbTransaction
	moduleName  string
	permissions ContextPermissions
	user        dmodel.DynamicFields
}

func (this RequestContext) InnerContext() context.Context {
	return this.Context
}

func (this RequestContext) GetLogger() logging.LoggerService {
	return this.logger
}

func (this *RequestContext) SetLogger(logger logging.LoggerService) {
	this.logger = logger
}

func (this RequestContext) GetDbTranx() db.DbTransaction {
	return this.repoTrx
}

func (this *RequestContext) SetDbTranx(trx db.DbTransaction) {
	this.repoTrx = trx
}

func (this RequestContext) GetDomainConstraints() dmodel.DynamicFields {
	val := this.Context.Value(CtxKeyDomainConstraints)
	if val == nil {
		return nil
	}
	return val.(dmodel.DynamicFields)
}

func (this *RequestContext) SetDomainConstraints(constraints dmodel.DynamicFields) {
	this.Context = context.WithValue(this.Context, CtxKeyDomainConstraints, constraints)
}

func (this RequestContext) GetPermissions() ContextPermissions {
	return this.permissions
}

func (this *RequestContext) SetPermissions(permissions ContextPermissions) {
	this.permissions = permissions
}

func (this RequestContext) GetUser() dmodel.DynamicFields {
	return this.user
}

func (this *RequestContext) SetUser(user dmodel.DynamicFields) {
	this.user = user
}

func (this RequestContext) GetModuleName() string {
	return this.moduleName
}

// Replace current inner context with a new one that has the given key and value.
func (this *RequestContext) WithValue(key, val any) {
	this.Context = context.WithValue(this.Context, key, val)
}

func (this *RequestContext) Value(key any) any {
	return this.Context.Value(key)
}
