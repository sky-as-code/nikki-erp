package grant_response

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type GrantResponseRepository interface {
	Create(ctx crud.Context, grantResponse domain.GrantResponse) (*domain.GrantResponse, error)
	FindByRequestId(ctx crud.Context, requestId model.Id) ([]domain.GrantResponse, error)
}
