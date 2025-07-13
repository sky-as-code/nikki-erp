package cqrs

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

type ServiceHandleFn[TRequest Request, TResult any] func(ctx context.Context, cmd TRequest) (*TResult, error)

func HandlePacket[TRequest Request, TResult any](ctx context.Context, packet *RequestPacket[TRequest], handleFn ServiceHandleFn[TRequest, TResult]) (*Reply[TResult], error) {
	cmd := packet.Request()
	result, err := handleFn(ctx, *cmd)
	ft.PanicOnErr(err)

	return &Reply[TResult]{
		Result: *result,
	}, nil
}
