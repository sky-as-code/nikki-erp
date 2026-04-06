package app

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
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

func (this *subjectHelper) assertSubjectExists(
	ctx context.Context,
	subjectType domain.SubjectType,
	subjectRef *model.Id,
	username *string,
	clientErrs *ft.ClientErrors,
) (subject *loginSubject, err error) {
	switch subjectType {
	case domain.SubjectTypeUser:
		subject, err = this.assertUserExists(ctx, subjectRef, username, clientErrs)
	}
	if err != nil {
		return nil, err
	}
	return subject, nil
}

func (this *subjectHelper) assertUserExists(
	ctx context.Context,
	userId *model.Id,
	username *string,
	clientErrs *ft.ClientErrors,
) (*loginSubject, error) {
	result := itUser.GetUserResult{}
	var field string
	var userIdStr *string
	if userId != nil {
		field = "id"
		userIdStr = util.ToPtr(string(*userId))
	} else {
		field = "email"
	}
	err := this.cqrsBus.Request(ctx, itUser.GetUserQuery{
		Email: username,
		Id:    userIdStr,
	}, &result)
	if err != nil {
		return nil, err
	}

	if !result.HasData {
		appendValidationError(clientErrs, "username", "user id not found or not active")
		return nil, nil
	}
	// If not validation error but another client error
	// if !vErrs.MergeClientError(result.ClientError) {
	// 	return nil, result.ClientError
	// }
	if result.ClientErrors.Count() > 0 {
		for i := range result.ClientErrors {
			item := result.ClientErrors[i]
			if item.Field == field {
				item.Field = "username"
			}
			clientErrs.Append(item)
		}
		return nil, nil
	}
	return &loginSubject{
		Id:       *result.Data.GetId(),
		Name:     *result.Data.GetDisplayName(),
		Username: *result.Data.GetEmail(),
	}, nil
}
