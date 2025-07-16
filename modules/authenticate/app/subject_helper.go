package app

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	domIdent "github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type subjectHelper struct {
	cqrsBus cqrs.CqrsBus
}

type loginSubject struct {
	Id       model.Id
	Name     string
	Username string
}

func (this *subjectHelper) assertSubjectExists(ctx context.Context, subjectType domain.SubjectType, username string, vErrs *ft.ValidationErrors) (subject *loginSubject, err error) {
	switch subjectType {
	case domain.SubjectTypeUser:
		subject, err = this.assertUserExists(ctx, username, vErrs)
	}
	if err != nil {
		return nil, err
	}
	return subject, nil
}

func (this *subjectHelper) assertUserExists(ctx context.Context, username string, vErrs *ft.ValidationErrors) (*loginSubject, error) {
	result := itUser.GetUserByEmailResult{}
	err := this.cqrsBus.Request(ctx, &itUser.GetUserByEmailQuery{
		Email:  username,
		Status: util.ToPtr(domIdent.UserStatusActive),
	}, &result)
	if err != nil {
		return nil, err
	}
	// If not validation error but another client error
	if !vErrs.MergeClientError(result.ClientError) {
		return nil, result.ClientError
	}
	if vErrs.Count() > 0 {
		vErrs.RenameKey("email", "username")
		return nil, nil
	}
	return &loginSubject{
		Id:       *result.Data.Id,
		Name:     *result.Data.DisplayName,
		Username: *result.Data.Email,
	}, nil
}
