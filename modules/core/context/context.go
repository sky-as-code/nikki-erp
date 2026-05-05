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

type ContextPermissions struct {
	IsOwner      bool
	Entitlements ds.Set[string]
	UserId       model.Id `json:"user_id"`
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

func NewRequestContextF(ctx context.Context, moduleName string, domainConstraints dmodel.DynamicFields) Context {
	return &RequestContext{
		Context:           ctx,
		domainConstraints: domainConstraints,
		moduleName:        moduleName,
	}
}

func CloneRequestContext(ctx Context) Context {
	// var dbTrx db.DbTransaction
	// if ctx.GetDbTranx() != nil {
	// 	dbTrx = ctx.GetDbTranx().(db.DbTransaction)
	// }
	return &RequestContext{
		Context: ctx.InnerContext(),
		logger:  ctx.GetLogger(),
		repoTrx: ctx.GetDbTranx(),
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
	repoTrx           db.DbTransaction
	domainConstraints dmodel.DynamicFields
	moduleName        string
	permissions       ContextPermissions
	userId            model.Id
	user              dmodel.DynamicFields
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