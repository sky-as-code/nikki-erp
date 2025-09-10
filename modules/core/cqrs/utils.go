package cqrs

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type ServiceHandleFn[TRequest Request, TResult any] func(ctx crud.Context, cmd TRequest) (*TResult, error)

func HandlePacket[TRequest Request, TResult any](ctx context.Context, packet *RequestPacket[TRequest], handleFn ServiceHandleFn[TRequest, TResult]) (*Reply[TResult], error) {
	cmd := packet.Request()
	reqCtx := crud.NewRequestContext(ctx)
	result, err := handleFn(reqCtx, *cmd)
	ft.PanicOnErr(err)

	return &Reply[TResult]{
		Result: *result,
	}, nil
}
