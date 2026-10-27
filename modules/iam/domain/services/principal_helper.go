package services

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

type principalHelper struct {
	cqrsBus cqrs.CqrsBus
	userSvc itUser.UserCrudDomainService
}

type loginPrincipal struct {
	Id       model.Id
	Name     string
	Username string
	Type     models.UserType
}

// userTypeOf is the type the principal's tokens carry. A principal that is not an IAM user, or a
// user row read before the field existed, is a person.
func userTypeOf(principal *loginPrincipal) models.UserType {
	if principal == nil || principal.Type == "" {
		return models.UserTypeUser
	}
	return principal.Type
}

func (this *principalHelper) assertPrincipalExists(
	ctx corectx.Context, principalType models.PrincipalType, principalId *model.Id,
	username *string, cErrs *ft.ClientErrors,
) (*loginPrincipal, error) {
	if principalType == models.PrincipalTypeNikkiUser {
		return this.assertUserExists(ctx, principalId, username, cErrs)
	}
	return nil, nil
}

func (this *principalHelper) assertUserExists(
	ctx corectx.Context, userId *model.Id, email *string, clientErrs *ft.ClientErrors,
) (*loginPrincipal, error) {
	var errField string
	query := itUser.GetUserQuery{}
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
			ft.ErrorKey("err_account_not_found", "iam"),
			"Account not found.",
		))
		return nil, nil
	}

	user := result.Data
	status := *user.GetStatus()
	if status != models.UserStatusInvited && status != models.UserStatusActive {
		clientErrs.Append(*ft.NewBusinessViolation(
			errField,
			ft.ErrorKey("err_account_not_active", "iam"),
			"Account not active.",
		))
		return nil, nil
	}

	principal := &loginPrincipal{
		Id:       result.Data.MustGetId(),
		Name:     result.Data.MustGetDisplayName(),
		Username: result.Data.MustGetEmail(),
	}
	if userType := result.Data.GetType(); userType != nil {
		principal.Type = *userType
	}
	return principal, nil
}
