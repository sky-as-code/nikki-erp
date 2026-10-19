package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

// GetUser reads one user by id or email, archived or not. It asserts nothing: it is the sign-in
// and principal-resolution path's read, which runs before there is a caller to authorize.
func (this *UserCrudDomainServiceImpl) GetUser(
	ctx corectx.Context, query itUser.GetUserQuery,
) (*itUser.GetUserDomainResult, error) {
	return this.getUserWithArchived(ctx, query, nil)
}

// GetEnabledUser is GetUser restricted to unarchived users: the request guard's read, through
// which an archived user stops being able to act.
func (this *UserCrudDomainServiceImpl) GetEnabledUser(
	ctx corectx.Context, query itUser.GetUserQuery,
) (*itUser.GetUserDomainResult, error) {
	return this.getUserWithArchived(ctx, query, util.ToPtr(false))
}

func (this *UserCrudDomainServiceImpl) getUserWithArchived(
	ctx corectx.Context, query itUser.GetUserQuery, isArchived *bool,
) (*itUser.GetUserDomainResult, error) {
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itUser.GetUserDomainResult{ClientErrors: cErrs}, nil
	}
	query = *(sanitized.(*itUser.GetUserQuery))

	statusNode := dmodel.NewSearchNode()
	if isArchived != nil {
		statusNode.NewCondition(basemodel.FieldIsArchived, dmodel.Equals, *isArchived)
	}
	keyNode := dmodel.NewSearchNode()
	if query.Id != nil {
		keyNode.NewCondition(domain.UserFieldId, dmodel.Equals, *query.Id)
	}
	if query.Email != nil {
		keyNode.NewCondition(domain.UserFieldEmail, dmodel.Equals, *query.Email)
	}
	graph := &dmodel.SearchGraph{}
	graph.And(*statusNode, *keyNode)

	searchRes, err := corecrud.Search[domain.User](ctx, corecrud.SearchParam{
		Action:       "get user",
		DbRepoGetter: this.Repository(),
		Query:        dyn.SearchQuery{Fields: query.Fields, Graph: graph, Page: 0, Size: 1},
	})
	if err != nil {
		return nil, err
	}
	result := &itUser.GetUserDomainResult{ClientErrors: searchRes.ClientErrors, HasData: searchRes.HasData}
	if searchRes.HasData {
		result.Data = searchRes.Data.Items[0]
	}
	return result, nil
}
