package identity_cqrs

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

type IdentityCqrsAdapter interface {
	UserExists(ctx context.Context, userId model.Id) (bool, error, *fault.ClientError)
}

func NewIdentityCqrsAdapter(cqrsBus cqrs.CqrsBus) IdentityCqrsAdapter {
	return &identityCqrsAdapter{
		cqrsBus: cqrsBus,
	}
}

type identityCqrsAdapter struct {
	cqrsBus cqrs.CqrsBus
}

func (this *identityCqrsAdapter) UserExists(ctx context.Context, userId model.Id) (bool, error, *fault.ClientError) {
	existCmd := UserExistsQuery{
		Id: userId,
	}
	existRes := UserExistsResult{}

	err := this.cqrsBus.Request(ctx, existCmd, &existRes)
	if err != nil {
		return false, err, nil
	}

	if existRes.ClientError != nil {
		return false, nil, existRes.ClientError
	}

	return existRes.HasData && existRes.Data, nil, nil
}
