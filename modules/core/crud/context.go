package crud

import (
	"context"

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
}

func NewRequestContext(ctx context.Context) Context {
	return &RequestContext{
		Context: ctx,
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

type RequestContext struct {
	context.Context

	logger logging.LoggerService

	// The transaction object that Repository Layer can use to perform atomic database operations.
	repoTrx db.DbTransaction
}

func (this *RequestContext) InnerContext() context.Context {
	return this.Context
}

func (this *RequestContext) GetLogger() logging.LoggerService {
	return this.logger
}

func (this *RequestContext) SetLogger(logger logging.LoggerService) {
	this.logger = logger
}

func (this *RequestContext) GetDbTranx() db.DbTransaction {
	return this.repoTrx
}

func (this *RequestContext) SetDbTranx(trx db.DbTransaction) {
	this.repoTrx = trx
}
