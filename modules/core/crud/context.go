package crud

import (
	"context"

	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type Context interface {
	context.Context

	GetLogger() logging.LoggerService
	SetLogger(logger logging.LoggerService)
	GetDbTranx() any
	SetDbTranx(trx db.DbTransaction)
}

func NewRequestContext(ctx context.Context) Context {
	return &RequestContext{
		Context: ctx,
	}
}

type RequestContext struct {
	context.Context

	logger logging.LoggerService

	// The transaction object that Repository Layer can use to perform atomic database operations.
	repoTrx db.DbTransaction
}

func (this *RequestContext) GetLogger() logging.LoggerService {
	return this.logger
}

func (this *RequestContext) SetLogger(logger logging.LoggerService) {
	this.logger = logger
}

func (this *RequestContext) GetDbTranx() any {
	return this.repoTrx
}

func (this *RequestContext) SetDbTranx(trx db.DbTransaction) {
	this.repoTrx = trx
}
