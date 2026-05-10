package escalationrule

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
)

type EscalationRuleRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.EscalationRule) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.EscalationRule) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, data models.EscalationRule) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.EscalationRule], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.EscalationRule]], error)
	Update(ctx corectx.Context, data models.EscalationRule) (*dyn.OpResult[dyn.MutateResultData], error)
}
