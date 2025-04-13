package utility

import (
	"context"
	"gorm.io/gorm"
)

func NewRequestContext() *RequestContext {
	return &RequestContext{}
}

type RequestContext struct {
	goContext       context.Context
	gormTransaction *gorm.DB
}

func (this *RequestContext) DbTx() *gorm.DB {
	return this.gormTransaction
}

func (this *RequestContext) GoContext() context.Context {
	return this.goContext
}

func (this *RequestContext) TODO() *RequestContext {
	this.goContext = context.TODO()
	return this
}

func (this *RequestContext) WithGoContext(ctx context.Context) *RequestContext {
	this.goContext = ctx
	return this
}

func (this *RequestContext) WithDbTx(tx *gorm.DB) *RequestContext {
	this.gormTransaction = tx
	return this
}

func (this *RequestContext) GetContext() context.Context {
	return this.goContext
}
