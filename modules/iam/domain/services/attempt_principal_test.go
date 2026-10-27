package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

type signinUserSvcStub struct {
	itUser.UserCrudDomainService
	user *models.User
}

func (this *signinUserSvcStub) GetUser(
	_ corectx.Context, _ itUser.GetUserQuery,
) (*itUser.GetUserDomainResult, error) {
	return &itUser.GetUserDomainResult{Data: *this.user, HasData: true}, nil
}

func signinUser(userType *models.UserType) *models.User {
	user := models.NewUser()
	user.SetId(util.ToPtr(model.Id("01JQZ0X0000000000000000001")))
	user.SetEmail(util.ToPtr("someone@example.com"))
	user.SetDisplayName(util.ToPtr("Someone"))
	user.SetStatus(util.ToPtr(models.UserStatusActive))
	user.SetType(userType)
	return user
}

func startSignin(t *testing.T, user *models.User) (*attemptPrincipal, ft.ClientErrors) {
	svc := &AttemptDomainServiceImpl{userSvc: &signinUserSvcStub{user: user}}
	cErrs := ft.ClientErrors{}
	principal, err := svc.assertUserExists(corectx.NewRequestContext(context.Background()), "someone@example.com", &cErrs)
	require.NoError(t, err)
	return principal, cErrs
}

func TestSigninStartAcceptsAHumanUser(t *testing.T) {
	for _, userType := range []*models.UserType{util.ToPtr(models.UserTypeUser), nil} {
		principal, cErrs := startSignin(t, signinUser(userType))
		assert.Zero(t, cErrs.Count())
		assert.NotNil(t, principal)
	}
}
