package client

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

// IdentityCqrsClient sends CQRS requests to the identity module using drive-local DTOs
// that mirror identity contracts (for future microservice boundary).
type IdentityCqrsClient interface {
	UserExists(ctx context.Context, query UserExistsQuery) (*UserExistsResult, error)
	SearchUsers(ctx context.Context, query SearchUsersQuery) (*SearchUsersResult, error)
	GetUserById(ctx context.Context, query GetUserByIdQuery) (*GetUserByIdResult, error)
}

func NewIdentityCqrsClient(bus cqrs.CqrsBus) IdentityCqrsClient {
	return &identityCqrsClient{cqrsBus: bus}
}

type identityCqrsClient struct {
	cqrsBus cqrs.CqrsBus
}

func (this *identityCqrsClient) UserExists(ctx context.Context, query UserExistsQuery) (*UserExistsResult, error) {
	var res UserExistsResult
	err := this.cqrsBus.Request(ctx, query, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (this *identityCqrsClient) SearchUsers(ctx context.Context, query SearchUsersQuery) (*SearchUsersResult, error) {
	var res SearchUsersResult
	err := this.cqrsBus.Request(ctx, query, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (this *identityCqrsClient) GetUserById(ctx context.Context, query GetUserByIdQuery) (*GetUserByIdResult, error) {
	var res GetUserByIdResult
	err := this.cqrsBus.Request(ctx, query, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
