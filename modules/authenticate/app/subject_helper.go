package app

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
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
	vErrs *ft.ValidationErrors,
) (subject *loginSubject, err error) {
	switch subjectType {
	case domain.SubjectTypeUser:
		subject, err = this.assertUserExists(ctx, subjectRef, username, vErrs)
	}
	if err != nil {
		return nil, err
	}
	return subject, nil
}

func (this *subjectHelper) assertUserExists(ctx context.Context, userId *string, username *string, vErrs *ft.ValidationErrors) (*loginSubject, error) {
	result := itUser.MustGetActiveUserResult{}
	var field string
	if userId != nil {
		field = "id"
	} else {
		field = "email"
	}
	err := this.cqrsBus.Request(ctx, &itUser.MustGetActiveUserQuery{
		Email: username,
		Id:    userId,
	}, &result)
	if err != nil {
		return nil, err
	}

	if result.Data == nil {
		vErrs.Append("user: ", "user id not found or not active")
		return nil, nil
	}
	// If not validation error but another client error
	// if !vErrs.MergeClientError(result.ClientError) {
	// 	return nil, result.ClientError
	// }
	if vErrs.Count() > 0 {
		// E.g: From {"email": "user is archived"} to {"username": "user is archived"}
		vErrs.RenameKey(field, "username")
		return nil, nil
	}
	return &loginSubject{
		Id:       *result.Data.Id,
		Name:     *result.Data.DisplayName,
		Username: *result.Data.Email,
	}, nil
}
