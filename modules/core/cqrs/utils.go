package cqrs

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type ServiceHandleFn[TRequest Request, TResult any] func(ctx corecrud.Context, cmd TRequest) (*TResult, error)
type ServiceHandleFn2[TRequest Request, TResult any] func(ctx corectx.Context, cmd TRequest) (*TResult, error)

func HandlePacket[TRequest Request, TResult any](ctx context.Context, packet *RequestPacket[TRequest], handleFn ServiceHandleFn[TRequest, TResult]) (*Reply[TResult], error) {
	cmd := packet.Request()
	reqCtx := corecrud.NewRequestContext(ctx)
	result, err := handleFn(reqCtx, *cmd)
	ft.PanicOnErr(err)

	return &Reply[TResult]{
		Result: *result,
	}, nil
}

func HandlePacket2[TRequest Request, TResult any](ctx context.Context, moduleName string, packet *RequestPacket[TRequest], handleFn ServiceHandleFn2[TRequest, TResult]) (*Reply[TResult], error) {
	cmd := packet.Request()
	reqCtx := corectx.NewRequestContextF(ctx, moduleName, nil)
	result, err := handleFn(reqCtx, *cmd)
	ft.PanicOnErr(err)

	return &Reply[TResult]{
		Result: *result,
	}, nil
}
