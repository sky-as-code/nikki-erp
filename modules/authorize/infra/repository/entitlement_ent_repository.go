package repository

// import (
// 	"context"

// 	"github.com/sky-as-code/nikki-erp/common/crud"
// 	"github.com/sky-as-code/nikki-erp/common/model"
// 	"github.com/sky-as-code/nikki-erp/common/orm"
// 	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// 	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
// 	entEff "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/effectiveentitlement"
// 	entEntmt "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlement"
// 	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/entitlement"
// )

// func NewEntitlementEntRepository(client *ent.Client) it.EntitlementRepository {
// 	return &EntitlementEntRepository{
// 		client: client,
// 	}
// }

// type EntitlementEntRepository struct {
// 	client *ent.Client
// }

// // func (this *EntitlementEntRepository) getUserEffectiveEntitlements(ctx context.Context, subject domain.Subject) ([]domain.Entitlement, error) {
// func (this *EntitlementEntRepository) getUserEffectiveEntitlements(ctx context.Context, userId model.Id) ([]domain.Entitlement, error) {
// 	effectiveEnts, err := this.client.EffectiveEntitlement.
// 		Query().
// 		Where(entEff.UserIDEQ(userId.String())).
// 		All(ctx)
// }
