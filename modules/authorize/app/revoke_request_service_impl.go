package app

import (
	"strings"

	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itRevokeRequest "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/revoke_request"
	itRole "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/role"
	itSuite "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/role_suite"
	itGroup "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewRevokeRequestServiceImpl(
	revokeRequestRepo itRevokeRequest.RevokeRequestRepository,
	roleRepo itRole.RoleRepository,
	suiteRepo itSuite.RoleSuiteRepository,
	cqrsBus cqrs.CqrsBus,
) itRevokeRequest.RevokeRequestService {
	return &RevokeRequestServiceImpl{
		revokeRequestRepo: revokeRequestRepo,
		roleRepo:          roleRepo,
		suiteRepo:         suiteRepo,
		cqrsBus:           cqrsBus,
	}
}

type RevokeRequestServiceImpl struct {
	revokeRequestRepo itRevokeRequest.RevokeRequestRepository
	roleRepo          itRole.RoleRepository
	suiteRepo         itSuite.RoleSuiteRepository
	cqrsBus           cqrs.CqrsBus
}

func (this *RevokeRequestServiceImpl) Create(ctx crud.Context, cmd itRevokeRequest.CreateRevokeRequestCommand) (*itRevokeRequest.CreateRevokeRequestResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.RevokeRequest, itRevokeRequest.CreateRevokeRequestCommand, itRevokeRequest.CreateRevokeRequestResult]{
		Action:              "create revoke request",
		Command:             cmd,
		AssertBusinessRules: this.assertExistAccess,
		RepoCreate:          this.processCreateRevokeRequest,
		SetDefault:          this.setRevokeRequestDefaults,
		Sanitize:            this.sanitizeRevokeRequest,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRevokeRequest.CreateRevokeRequestResult {
			return &itRevokeRequest.CreateRevokeRequestResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.RevokeRequest) *itRevokeRequest.CreateRevokeRequestResult {
			return &itRevokeRequest.CreateRevokeRequestResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *RevokeRequestServiceImpl) CreateBulk(ctx crud.Context, cmd itRevokeRequest.CreateBulkRevokeRequestsCommand) (*itRevokeRequest.CreateBulkRevokeRequestsResult, error) {
	// Guardrail: don't allow empty bulk requests (crud.CreateBulk would treat it as success otherwise).
	if len(cmd.Items) == 0 {
		vErrs := fault.NewValidationErrors()
		vErrs.Append("items", "items must not be empty")
		return &itRevokeRequest.CreateBulkRevokeRequestsResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	result, err := crud.CreateBulk(ctx, crud.CreateBulkParam[*domain.RevokeRequest, itRevokeRequest.CreateBulkRevokeRequestsCommand, itRevokeRequest.CreateBulkRevokeRequestsResult]{
		Action:              "create bulk revoke requests",
		Command:             cmd,
		AssertBusinessRules: this.assertExistAccess,
		RepoCreateBulk:      this.processCreateRevokeRequests,
		SetDefault:          this.setRevokeRequestDefaults,
		Sanitize:            this.sanitizeRevokeRequest,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRevokeRequest.CreateBulkRevokeRequestsResult {
			return &itRevokeRequest.CreateBulkRevokeRequestsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(models []*domain.RevokeRequest) *itRevokeRequest.CreateBulkRevokeRequestsResult {
			return &itRevokeRequest.CreateBulkRevokeRequestsResult{
				Data:    models,
				HasData: models != nil,
			}
		},
	})

	return result, err
}

func (this *RevokeRequestServiceImpl) TargetIsDeleted(ctx crud.Context, cmd itRevokeRequest.TargetIsDeletedCommand) (result *itRevokeRequest.TargetIsDeletedResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "target deleted"); e != nil {
			err = e
		}
	}()

	var revokeRequests []domain.RevokeRequest

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			revokeRequests, err = this.findRevokeRequestsByTarget(ctx, cmd.TargetType, cmd.TargetRef)
			return err
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itRevokeRequest.TargetIsDeletedResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	for _, revokeRequest := range revokeRequests {
		prevEtag := revokeRequest.Etag
		revokeRequest.Etag = model.NewEtag()

		targetType := domain.GrantRequestTargetType(cmd.TargetType)
		switch targetType {
		case domain.GrantRequestTargetTypeRole:
			revokeRequest.TargetRoleName = &cmd.TargetName
		case domain.GrantRequestTargetTypeSuite:
			revokeRequest.TargetSuiteName = &cmd.TargetName
		}

		err = this.revokeRequestRepo.UpdateTargetFields(ctx, &revokeRequest, *prevEtag)
		fault.PanicOnErr(err)
	}

	return &itRevokeRequest.TargetIsDeletedResult{
		Data:    true,
		HasData: false,
	}, nil
}

func (this *RevokeRequestServiceImpl) Delete(ctx crud.Context, cmd itRevokeRequest.DeleteRevokeRequestCommand) (*itRevokeRequest.DeleteRevokeRequestResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.RevokeRequest, itRevokeRequest.DeleteRevokeRequestCommand, itRevokeRequest.DeleteRevokeRequestResult]{
		Action:       "delete revoke request",
		Command:      cmd,
		AssertExists: this.assertRevokeRequestExists,
		RepoDelete: func(ctx crud.Context, model *domain.RevokeRequest) (int, error) {
			return this.revokeRequestRepo.Delete(ctx, itRevokeRequest.DeleteParam{Id: *model.Id})
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRevokeRequest.DeleteRevokeRequestResult {
			return &itRevokeRequest.DeleteRevokeRequestResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.RevokeRequest, deletedCount int) *itRevokeRequest.DeleteRevokeRequestResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})

	return result, err
}

func (this *RevokeRequestServiceImpl) GetById(ctx crud.Context, query itRevokeRequest.GetRevokeRequestByIdQuery) (*itRevokeRequest.GetRevokeRequestByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.RevokeRequest, itRevokeRequest.GetRevokeRequestByIdQuery, itRevokeRequest.GetRevokeRequestByIdResult]{
		Action:      "get revoke request by id",
		Query:       query,
		RepoFindOne: this.getRevokeRequestById,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRevokeRequest.GetRevokeRequestByIdResult {
			return &itRevokeRequest.GetRevokeRequestByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.RevokeRequest) *itRevokeRequest.GetRevokeRequestByIdResult {
			return &itRevokeRequest.GetRevokeRequestByIdResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *RevokeRequestServiceImpl) Search(ctx crud.Context, query itRevokeRequest.SearchRevokeRequestsQuery) (*itRevokeRequest.SearchRevokeRequestsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.RevokeRequest, itRevokeRequest.SearchRevokeRequestsQuery, itRevokeRequest.SearchRevokeRequestsResult]{
		Action: "search revoke requests",
		Query:  query,
		SetQueryDefaults: func(query *itRevokeRequest.SearchRevokeRequestsQuery) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.revokeRequestRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query itRevokeRequest.SearchRevokeRequestsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.RevokeRequest], error) {
			result, err := this.revokeRequestRepo.Search(ctx, itRevokeRequest.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})

			for i := range result.Items {
				err := this.populateRevokeRequestDetails(ctx, &result.Items[i])
				fault.PanicOnErr(err)
			}

			return result, err
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itRevokeRequest.SearchRevokeRequestsResult {
			return &itRevokeRequest.SearchRevokeRequestsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.RevokeRequest]) *itRevokeRequest.SearchRevokeRequestsResult {
			return &itRevokeRequest.SearchRevokeRequestsResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
}

func (this *RevokeRequestServiceImpl) setRevokeRequestDefaults(revokeRequest *domain.RevokeRequest) {
	revokeRequest.SetDefaults()
}

func (this *RevokeRequestServiceImpl) assertExistAccess(ctx crud.Context, revokeRequest *domain.RevokeRequest, vErrs *fault.ValidationErrors) error {
	if *revokeRequest.TargetType == domain.RevokeRequestTargetTypeNikkiRole {
		exist, err := this.roleRepo.ExistUserWithRole(ctx, itRole.ExistUserWithRoleParam{
			TargetId:     *revokeRequest.TargetRef,
			ReceiverType: this.classifyReceiverType(revokeRequest.ReceiverType),
			ReceiverId:   *revokeRequest.ReceiverId,
		})
		fault.PanicOnErr(err)

		if !exist {
			vErrs.AppendNotFound("target", "target role user")
		}
	} else {
		existSuiteUser, err := this.suiteRepo.ExistUserWithRoleSuite(ctx, itSuite.ExistUserWithRoleSuiteParam{
			TargetId:     *revokeRequest.TargetRef,
			ReceiverType: this.classifyReceiverType(revokeRequest.ReceiverType),
			ReceiverId:   *revokeRequest.ReceiverId,
		})
		fault.PanicOnErr(err)

		if !existSuiteUser {
			vErrs.AppendNotFound("target", "target suite user")
		}
	}

	return nil
}

func (this *RevokeRequestServiceImpl) sanitizeRevokeRequest(revokeRequest *domain.RevokeRequest) {
	if revokeRequest.Comment != nil {
		cleanedComment := strings.TrimSpace(*revokeRequest.Comment)
		cleanedComment = defense.SanitizePlainText(cleanedComment)
		revokeRequest.Comment = &cleanedComment
	}

	if revokeRequest.AttachmentURL != nil {
		cleanedAttachmentUrl := strings.TrimSpace(*revokeRequest.AttachmentURL)
		cleanedAttachmentUrl = defense.SanitizePlainText(cleanedAttachmentUrl)
		revokeRequest.AttachmentURL = &cleanedAttachmentUrl
	}
}

func (this *RevokeRequestServiceImpl) processCreateRevokeRequest(ctx crud.Context, revokeRequest *domain.RevokeRequest) (*domain.RevokeRequest, error) {
	tx, err := this.revokeRequestRepo.BeginTransaction(ctx)
	fault.PanicOnErr(err)

	ctx.SetDbTranx(tx)

	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "transaction process revoke request"); e != nil {
			err = e
			tx.Rollback()
			return
		}

		tx.Commit()
	}()

	result, err := this.revokeRequestRepo.Create(ctx, revokeRequest)
	fault.PanicOnErr(err)

	if *revokeRequest.TargetType == domain.RevokeRequestTargetTypeNikkiRole {
		err := this.roleRepo.AddRemoveUser(ctx, itRole.AddRemoveUserParam{
			Id:           *revokeRequest.TargetRef,
			ReceiverID:   *revokeRequest.ReceiverId,
			ReceiverType: this.classifyReceiverType(revokeRequest.ReceiverType),
			Add:          false,
		})
		fault.PanicOnErr(err)
	} else {
		err := this.suiteRepo.AddRemoveUser(ctx, itSuite.AddRemoveUserParam{
			Id:           *revokeRequest.TargetRef,
			ReceiverID:   *revokeRequest.ReceiverId,
			ReceiverType: this.classifyReceiverType(revokeRequest.ReceiverType),
			Add:          false,
		})
		fault.PanicOnErr(err)
	}

	return result, err
}

func (this *RevokeRequestServiceImpl) processCreateRevokeRequests(ctx crud.Context, revokeRequests []*domain.RevokeRequest) ([]*domain.RevokeRequest, error) {
	tx, err := this.revokeRequestRepo.BeginTransaction(ctx)
	fault.PanicOnErr(err)

	ctx.SetDbTranx(tx)

	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "transaction process bulk revoke requests"); e != nil {
			err = e
			_ = tx.Rollback()
			return
		}

		if err != nil {
			_ = tx.Rollback()
			return
		}
		_ = tx.Commit()
	}()

	created, err := this.revokeRequestRepo.CreateBulk(ctx, revokeRequests)
	fault.PanicOnErr(err)

	// Apply revokes (remove assignments) inside the same transaction.
	for _, revokeRequest := range revokeRequests {
		if *revokeRequest.TargetType == domain.RevokeRequestTargetTypeNikkiRole {
			err := this.roleRepo.AddRemoveUser(ctx, itRole.AddRemoveUserParam{
				Id:           *revokeRequest.TargetRef,
				ReceiverID:   *revokeRequest.ReceiverId,
				ReceiverType: this.classifyReceiverType(revokeRequest.ReceiverType),
				Add:          false,
			})
			fault.PanicOnErr(err)
		} else {
			err := this.suiteRepo.AddRemoveUser(ctx, itSuite.AddRemoveUserParam{
				Id:           *revokeRequest.TargetRef,
				ReceiverID:   *revokeRequest.ReceiverId,
				ReceiverType: this.classifyReceiverType(revokeRequest.ReceiverType),
				Add:          false,
			})
			fault.PanicOnErr(err)
		}
	}

	return created, err
}

func (this *RevokeRequestServiceImpl) classifyReceiverType(targetType *domain.ReceiverType) domain.ReceiverType {
	if *targetType == domain.ReceiverTypeUser {
		return domain.ReceiverTypeUser
	} else {
		return domain.ReceiverTypeGroup
	}
}

func (this *RevokeRequestServiceImpl) getRevokeRequestById(ctx crud.Context, query itRevokeRequest.GetRevokeRequestByIdQuery, vErrs *fault.ValidationErrors) (dbRevokeRequest *domain.RevokeRequest, err error) {
	dbRevokeRequest, err = this.revokeRequestRepo.FindById(ctx, itRevokeRequest.FindByIdParam{Id: query.Id})
	fault.PanicOnErr(err)

	if dbRevokeRequest == nil {
		vErrs.AppendNotFound("id", "revoke request id")
		return
	}

	err = this.populateRevokeRequestDetails(ctx, dbRevokeRequest)
	fault.PanicOnErr(err)

	return
}

func (this *RevokeRequestServiceImpl) populateRevokeRequestDetails(ctx crud.Context, dbRevokeRequest *domain.RevokeRequest) (err error) {
	dbRevokeRequest.RequestorName, err = this.getUserDisplayName(ctx, *dbRevokeRequest.RequestorId, "user", nil)
	fault.PanicOnErr(err)

	if *dbRevokeRequest.ReceiverType == domain.ReceiverTypeGroup {
		dbRevokeRequest.ReceiverName, err = this.getUserDisplayName(ctx, *dbRevokeRequest.ReceiverId, "group", nil)
	} else {
		dbRevokeRequest.ReceiverName, err = this.getUserDisplayName(ctx, *dbRevokeRequest.ReceiverId, "user", nil)
	}
	fault.PanicOnErr(err)

	return
}

func (this *RevokeRequestServiceImpl) getUserDisplayName(ctx crud.Context, id model.Id, entityType string, vErrs *fault.ValidationErrors) (*string, error) {
	switch entityType {
	case "user":
		cmd := &itUser.GetUserByIdQuery{Id: id}
		res := itUser.GetUserByIdResult{}
		err := this.cqrsBus.Request(ctx, *cmd, &res)
		fault.PanicOnErr(err)

		if res.ClientError != nil {
			vErrs.MergeClientError(res.ClientError)
			return nil, nil
		}
		if res.Data == nil {
			return nil, nil
		}

		return res.Data.DisplayName, nil

	case "group":
		cmd := &itGroup.GetGroupByIdQuery{Id: id}
		res := itGroup.GetGroupByIdResult{}
		err := this.cqrsBus.Request(ctx, *cmd, &res)
		fault.PanicOnErr(err)

		if res.ClientError != nil {
			vErrs.MergeClientError(res.ClientError)
			return nil, nil
		}
		if res.Data == nil {
			return nil, nil
		}

		return res.Data.Name, nil
	}

	return nil, nil
}

func (this *RevokeRequestServiceImpl) assertRevokeRequestExists(ctx crud.Context, revokeRequest *domain.RevokeRequest, vErrs *fault.ValidationErrors) (dbRevokeRequest *domain.RevokeRequest, err error) {
	dbRevokeRequest, err = this.revokeRequestRepo.FindById(ctx, itRevokeRequest.FindByIdParam{Id: *revokeRequest.Id})
	fault.PanicOnErr(err)

	if dbRevokeRequest == nil {
		vErrs.AppendNotFound("id", "revoke request")
	}

	return
}

func (this *RevokeRequestServiceImpl) findRevokeRequestsByTarget(ctx crud.Context, targetType domain.GrantRequestTargetType, targetRef model.Id) ([]domain.RevokeRequest, error) {
	return this.revokeRequestRepo.FindAllByTarget(ctx, itRevokeRequest.FindAllByTargetParam{
		TargetType: targetType,
		TargetRef:  targetRef,
	})
}
