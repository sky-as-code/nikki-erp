package app

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	ext "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

type principalHelper struct {
	cqrsBus cqrs.CqrsBus
	userSvc ext.UserExtService
}

type loginPrincipal struct {
	Id       model.Id
	Name     string
	Username string
}

func (this *principalHelper) assertPrincipalExists(
	ctx corectx.Context, principalType domain.PrincipalType, principalId *model.Id,
	username *string, cErrs *ft.ClientErrors,
) (*loginPrincipal, error) {
	if principalType == domain.PrincipalTypeNikkiUser {
		return this.assertUserExists(ctx, principalId, username, cErrs)
	}
	return nil, nil
}

func (this *principalHelper) assertUserExists(
	ctx corectx.Context, userId *model.Id, email *string, clientErrs *ft.ClientErrors,
) (*loginPrincipal, error) {
	var errField string
	query := ext.GetUserQuery{}
	if userId != nil {
		query.Id = userId
		errField = "principal_id"
	}
	if email != nil {
		query.Email = email
		errField = "username"
	}
	result, err := this.userSvc.GetUser(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(result.ClientErrors.ToError(), "assertUserExists")
	}

	if !result.HasData {
		clientErrs.Append(*ft.NewBusinessViolation(
			errField,
			ft.ErrorKey("err_account_not_found", "authenticate"),
			"Account not found.",
		))
		return nil, nil
	}

	user := result.Data
	status := *user.GetStatus()
	if status != ext.UserStatusInvited && status != ext.UserStatusActive {
		clientErrs.Append(*ft.NewBusinessViolation(
			errField,
			ft.ErrorKey("err_account_not_active", "authenticate"),
			"Account not active.",
		))
		return nil, nil
	}

	return &loginPrincipal{
		Id:       result.Data.MustGetId(),
		Name:     result.Data.MustGetDisplayName(),
		Username: result.Data.MustGetEmail(),
	}, nil
}
