package app

import (
	"fmt"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_request"
	itGrantResponse "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_response"
	itPermissionHistory "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/permission_history"
	itRole "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
	itRoleSuite "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
	itGroup "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type ApprovalType int

const (
	ApprovalTypeManagerOnly ApprovalType = iota
	ApprovalTypeManagerAndOwner
	ApprovalTypeOwnerOnly
	ApprovalTypeNone
)

type ResponseState struct {
	AnyManagerResponded bool
	AnyManagerDenied    bool
	AnyManagerApproved  bool
	AnyOwnerDenied      bool
	AnyOwnerApproved    bool
}

func NewGrantRequestServiceImpl(
	grantRequestRepo itGrantRequest.GrantRequestRepository,
	grantResponseRepo itGrantResponse.GrantResponseRepository,
	roleRepo itRole.RoleRepository,
	suiteRepo itRoleSuite.RoleSuiteRepository,
	permissionHistoryRepo itPermissionHistory.PermissionHistoryRepository,
	cqrsBus cqrs.CqrsBus,
) itGrantRequest.GrantRequestService {
	return &GrantRequestServiceImpl{
		grantRequestRepo:      grantRequestRepo,
		grantResponseRepo:     grantResponseRepo,
		roleRepo:              roleRepo,
		suiteRepo:             suiteRepo,
		permissionHistoryRepo: permissionHistoryRepo,
		cqrsBus:               cqrsBus,
	}
}

type GrantRequestServiceImpl struct {
	grantRequestRepo      itGrantRequest.GrantRequestRepository
	grantResponseRepo     itGrantResponse.GrantResponseRepository
	roleRepo              itRole.RoleRepository
	suiteRepo             itRoleSuite.RoleSuiteRepository
	permissionHistoryRepo itPermissionHistory.PermissionHistoryRepository
	cqrsBus               cqrs.CqrsBus
}

func (this *GrantRequestServiceImpl) TargetIsDeleted(ctx crud.Context, cmd itGrantRequest.TargetIsDeletedCommand) (result *itGrantRequest.TargetIsDeletedResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "role id deleted"); e != nil {
			err = e
		}
	}()

	var grantRequests []domain.GrantRequest

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			grantRequests, err = this.findGrantRequestsByTarget(ctx, cmd.TargetType, cmd.TargetRef)
			return err
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itGrantRequest.TargetIsDeletedResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	for _, grantRequest := range grantRequests {
		prevEtag := grantRequest.Etag
		grantRequest.Etag = model.NewEtag()
		if grantRequest.Status != nil && *grantRequest.Status == domain.PendingGrantRequestStatus {
			cancelStatus := domain.CancelledGrantRequestStatus
			grantRequest.Status = &cancelStatus
		}

		err = this.grantRequestRepo.ConfigTargetFields(ctx, &grantRequest, cmd.TargetName, *prevEtag)
		fault.PanicOnErr(err)
	}

	return &itGrantRequest.TargetIsDeletedResult{
		Data:    true,
		HasData: false,
	}, nil
}

func (this *GrantRequestServiceImpl) CreateGrantRequest(ctx crud.Context, cmd itGrantRequest.CreateGrantRequestCommand) (result *itGrantRequest.CreateGrantRequestResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "create grant request"); e != nil {
			err = e
		}
	}()

	grantRequest := cmd.ToGrantRequest()
	this.setGrantRequestDefaults(grantRequest)

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.assertTarget(ctx, grantRequest, vErrs)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.assertReceiver(ctx, grantRequest, vErrs)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.assertReceiverNotAlreadyGranted(ctx, grantRequest, vErrs)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			return this.assertNoPendingGrantRequest(ctx, cmd, vErrs)
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			return this.assertOrgExists(ctx, grantRequest, vErrs)
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			return this.setupApprovalChain(ctx, grantRequest, vErrs)
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itGrantRequest.CreateGrantRequestResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	createdGrantRequest, err := this.grantRequestRepo.Create(ctx, grantRequest)
	fault.PanicOnErr(err)

	return &itGrantRequest.CreateGrantRequestResult{
		Data:    createdGrantRequest,
		HasData: createdGrantRequest != nil,
	}, nil
}

func (this *GrantRequestServiceImpl) CancelGrantRequest(ctx crud.Context, cmd itGrantRequest.CancelGrantRequestCommand) (result *itGrantRequest.CancelGrantRequestResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "cancel grant request"); e != nil {
			err = e
		}
	}()

	grantRequest := cmd.ToGrantRequest()
	var dbGrantRequest *domain.GrantRequest

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbGrantRequest, err = this.assertGrantRequestExists(ctx, grantRequest, vErrs)

			if dbGrantRequest != nil && *dbGrantRequest.Status != domain.PendingGrantRequestStatus {
				vErrs.Append("grant_request_id", "grant request is not pending")
			}
			return err
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.assertCorrectEtag(*grantRequest.Etag, *dbGrantRequest.Etag, vErrs)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			if *dbGrantRequest.RequestorId != cmd.ResponderId {
				vErrs.Append("responder_id", "not authorized to cancel this request")
			}
			return nil
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itGrantRequest.CancelGrantRequestResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := dbGrantRequest.Etag
	grantRequest.Etag = model.NewEtag()
	status := domain.CancelledGrantRequestStatus
	grantRequest.Status = &status
	update, err := this.grantRequestRepo.Update(ctx, grantRequest, *prevEtag)
	fault.PanicOnErr(err)

	return &itGrantRequest.CancelGrantRequestResult{
		Data:    update,
		HasData: true,
	}, nil
}

func (this *GrantRequestServiceImpl) DeleteGrantRequest(ctx crud.Context, cmd itGrantRequest.DeleteGrantRequestCommand) (*itGrantRequest.DeleteGrantRequestResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.GrantRequest, itGrantRequest.DeleteGrantRequestCommand, itGrantRequest.DeleteGrantRequestResult]{
		Action:       "delete grant request",
		Command:      cmd,
		AssertExists: this.assertGrantRequestExists,
		RepoDelete: func(ctx crud.Context, model *domain.GrantRequest) (int, error) {
			return this.processDeleteGrantRequest(ctx, model)
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itGrantRequest.DeleteGrantRequestResult {
			return &itGrantRequest.DeleteGrantRequestResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.GrantRequest, deletedCount int) *itGrantRequest.DeleteGrantRequestResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})

	return result, err
}

func (this *GrantRequestServiceImpl) RespondToGrantRequest(ctx crud.Context, cmd itGrantRequest.RespondToGrantRequestCommand) (result *itGrantRequest.RespondToGrantRequestResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "respond to grant request"); e != nil {
			err = e
		}
	}()

	grantRequest := cmd.ToGrantRequest()
	var dbGrantRequest *domain.GrantRequest
	var managerIds []string
	var ownerId *string

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbGrantRequest, err = this.assertGrantRequestExists(ctx, grantRequest, vErrs)

			if dbGrantRequest != nil && *dbGrantRequest.Status != domain.PendingGrantRequestStatus {
				vErrs.Append("grant_request_id", "grant request is not pending")
			}
			return err
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.assertCorrectEtag(cmd.Etag, *dbGrantRequest.Etag, vErrs)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			managerIds, ownerId, err = this.assertValidApprover(ctx, dbGrantRequest, cmd.ResponderId, vErrs)
			return err
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itGrantRequest.RespondToGrantRequestResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	tx, err := this.grantRequestRepo.BeginTransaction(ctx)
	fault.PanicOnErr(err)

	ctx.SetDbTranx(tx)

	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "transaction process grant response"); e != nil {
			err = e
			tx.Rollback()
			return
		}

		tx.Commit()
	}()

	prevEtag := dbGrantRequest.Etag
	dbGrantRequest.Etag = model.NewEtag()
	respondGrantRequest, err := this.processGrantResponse(ctx, dbGrantRequest, *prevEtag, cmd, managerIds, ownerId)
	fault.PanicOnErr(err)

	return &itGrantRequest.RespondToGrantRequestResult{
		Data:    respondGrantRequest,
		HasData: true,
	}, nil
}

func (this *GrantRequestServiceImpl) GetGrantRequestById(ctx crud.Context, query itGrantRequest.GetGrantRequestByIdQuery) (*itGrantRequest.GetGrantRequestByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.GrantRequest, itGrantRequest.GetGrantRequestByIdQuery, itGrantRequest.GetGrantRequestByIdResult]{
		Action:      "get grant request by Id",
		Query:       query,
		RepoFindOne: this.getGrantRequestById,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itGrantRequest.GetGrantRequestByIdResult {
			return &itGrantRequest.GetGrantRequestByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.GrantRequest) *itGrantRequest.GetGrantRequestByIdResult {
			return &itGrantRequest.GetGrantRequestByIdResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *GrantRequestServiceImpl) SearchGrantRequests(ctx crud.Context, query itGrantRequest.SearchGrantRequestsQuery) (*itGrantRequest.SearchGrantRequestsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.GrantRequest, itGrantRequest.SearchGrantRequestsQuery, itGrantRequest.SearchGrantRequestsResult]{
		Action: "search grant requests",
		Query:  query,
		SetQueryDefaults: func(query *itGrantRequest.SearchGrantRequestsQuery) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.grantRequestRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query itGrantRequest.SearchGrantRequestsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.GrantRequest], error) {
			result, err := this.grantRequestRepo.Search(ctx, itGrantRequest.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})

			// Collect unique orgIds for batch fetching
			orgIdSet := make(map[model.Id]bool)
			for i := range result.Items {
				if result.Items[i].OrgId != nil {
					orgIdSet[*result.Items[i].OrgId] = true
				}
			}

			// Batch fetch organizations
			orgMap := make(map[model.Id]*domain.Organization)
			for orgId := range orgIdSet {
				org, err := this.getOrganizationById(ctx, orgId)
				fault.PanicOnErr(err)
				if org != nil {
					orgMap[orgId] = org
				}
			}

			for i := range result.Items {
				err := this.populateGrantRequestDetails(ctx, &result.Items[i])
				fault.PanicOnErr(err)

				// Set organization from batch fetch
				if result.Items[i].OrgId != nil {
					if org, exists := orgMap[*result.Items[i].OrgId]; exists {
						result.Items[i].Organization = org
						result.Items[i].OrgName = org.DisplayName
					}
				}
			}

			return result, err
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *itGrantRequest.SearchGrantRequestsResult {
			return &itGrantRequest.SearchGrantRequestsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.GrantRequest]) *itGrantRequest.SearchGrantRequestsResult {
			return &itGrantRequest.SearchGrantRequestsResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
}

func (this *GrantRequestServiceImpl) setGrantRequestDefaults(grantRequest *domain.GrantRequest) {
	grantRequest.SetDefaults()
}

func (this *GrantRequestServiceImpl) assertOrgExists(ctx crud.Context, grantRequest *domain.GrantRequest, vErrs *fault.ValidationErrors) error {
	if grantRequest.OrgId == nil {
		return nil
	}

	existCmd := &itOrg.ExistsOrgByIdCommand{
		Id: *grantRequest.OrgId,
	}
	existRes := itOrg.ExistsOrgByIdResult{}
	err := this.cqrsBus.Request(ctx, *existCmd, &existRes)
	fault.PanicOnErr(err)

	if existRes.ClientError != nil {
		vErrs.MergeClientError(existRes.ClientError)
		return nil
	}

	if !existRes.Data {
		vErrs.Append("org_id", "not existing organization")
	}
	return nil
}

func (this *GrantRequestServiceImpl) assertTarget(ctx crud.Context, grantRequest *domain.GrantRequest, vErrs *fault.ValidationErrors) {
	var targetOrgId *model.Id

	switch *grantRequest.TargetType {
	case domain.GrantRequestTargetTypeRole:
		role, err := this.roleRepo.FindById(ctx, itRole.FindByIdParam{Id: *grantRequest.TargetRef})
		fault.PanicOnErr(err)

		if role == nil {
			vErrs.AppendNotFound("targetRef", "target")
			return
		}

		targetOrgId = role.OrgId
		this.validateTarget(role.IsRequestable, role.IsRequiredAttachment, role.IsRequiredComment, grantRequest, vErrs)
	case domain.GrantRequestTargetTypeSuite:
		suite, err := this.suiteRepo.FindById(ctx, itRoleSuite.FindByIdParam{Id: *grantRequest.TargetRef})
		fault.PanicOnErr(err)

		if suite == nil {
			vErrs.AppendNotFound("targetRef", "target")
			return
		}

		targetOrgId = suite.OrgId
		this.validateTarget(suite.IsRequestable, suite.IsRequiredAttachment, suite.IsRequiredComment, grantRequest, vErrs)
	}

	if targetOrgId != nil {
		grantRequest.OrgId = targetOrgId
	}
}

func (this *GrantRequestServiceImpl) assertReceiver(ctx crud.Context, grantRequest *domain.GrantRequest, vErrs *fault.ValidationErrors) {
	switch *grantRequest.ReceiverType {
	case domain.ReceiverTypeUser:
		existCmd := &itUser.UserExistsQuery{
			Id: *grantRequest.ReceiverId,
		}
		existRes := itUser.UserExistsResult{}
		err := this.cqrsBus.Request(ctx, *existCmd, &existRes)
		fault.PanicOnErr(err)

		if existRes.ClientError != nil {
			vErrs.MergeClientError(existRes.ClientError)
			return
		}

		if !existRes.Data {
			vErrs.Append("receiverId", "not existing user")
		}
		return
	case domain.ReceiverTypeGroup:
		existCmd := &itGroup.GroupExistsCommand{
			Id: *grantRequest.ReceiverId,
		}
		existRes := itGroup.GroupExistsResult{}
		err := this.cqrsBus.Request(ctx, *existCmd, &existRes)
		fault.PanicOnErr(err)

		if existRes.ClientError != nil {
			vErrs.MergeClientError(existRes.ClientError)
			return
		}

		if !existRes.Data {
			vErrs.Append("receiverId", "not existing group")
		}
		return
	}
}

func (this *GrantRequestServiceImpl) validateTarget(isRequestable, isRequiredAttachment, isRequiredComment *bool, grantRequest *domain.GrantRequest, vErrs *fault.ValidationErrors) {
	if isRequestable == nil || !*isRequestable {
		vErrs.Append("targetRef", "target is not requestable")
		return
	}

	if isRequiredAttachment != nil && *isRequiredAttachment {
		if grantRequest.AttachmentURL == nil || *grantRequest.AttachmentURL == "" {
			vErrs.Append("attachmentUrl", "attachment is required")
		}
	}

	if isRequiredComment != nil && *isRequiredComment {
		if grantRequest.Comment == nil || *grantRequest.Comment == "" {
			vErrs.Append("comment", "comment is required")
		}
	}
}

func (this *GrantRequestServiceImpl) assertReceiverNotAlreadyGranted(ctx crud.Context, grantRequest *domain.GrantRequest, vErrs *fault.ValidationErrors) {
	switch *grantRequest.TargetType {
	case domain.GrantRequestTargetTypeRole:
		exist, err := this.roleRepo.ExistUserWithRole(ctx, itRole.ExistUserWithRoleParam{
			ReceiverType: *grantRequest.ReceiverType,
			ReceiverId:   *grantRequest.ReceiverId,
			TargetId:     *grantRequest.TargetRef,
		})
		fault.PanicOnErr(err)

		if exist {
			vErrs.AppendAlreadyExists("receiverId", "granted receiver")
		}
	case domain.GrantRequestTargetTypeSuite:
		exist, err := this.suiteRepo.ExistUserWithRoleSuite(ctx, itRoleSuite.ExistUserWithRoleSuiteParam{
			ReceiverType: *grantRequest.ReceiverType,
			ReceiverId:   *grantRequest.ReceiverId,
			TargetId:     *grantRequest.TargetRef,
		})
		fault.PanicOnErr(err)

		if exist {
			vErrs.AppendAlreadyExists("receiverId", "granted receiver")
		}
	}
}

func (this *GrantRequestServiceImpl) assertNoPendingGrantRequest(ctx crud.Context, cmd itGrantRequest.CreateGrantRequestCommand, vErrs *fault.ValidationErrors) error {
	pendingRequests, err := this.grantRequestRepo.FindPendingByReceiverAndTarget(ctx, cmd.ReceiverId, cmd.TargetRef, domain.GrantRequestTargetType(cmd.TargetType))
	fault.PanicOnErr(err)

	if len(pendingRequests) > 0 {
		vErrs.AppendAlreadyExists("receiverId", "receiver already has a pending request for this role/suite already exists")
	}

	return nil
}

func (this *GrantRequestServiceImpl) setupApprovalChain(ctx crud.Context, grantRequest *domain.GrantRequest, vErrs *fault.ValidationErrors) error {
	status := domain.PendingGrantRequestStatus
	grantRequest.Status = &status

	notifyUserIds, message, err := this.determineInitialNotifications(ctx, grantRequest, vErrs)
	fault.PanicOnErr(err)

	for _, userId := range notifyUserIds {
		err = this.sendNotification(userId, message)
		fault.PanicOnErr(err)
	}

	return nil
}

func (this *GrantRequestServiceImpl) determineInitialNotifications(ctx crud.Context, grantRequest *domain.GrantRequest, vErrs *fault.ValidationErrors) ([]string, string, error) {
	ownerUserIds, err := this.findOwnerUserIds(ctx, *grantRequest.TargetRef, *grantRequest.TargetType, vErrs)
	fault.PanicOnErr(err)

	if *grantRequest.ReceiverType == domain.ReceiverTypeGroup {
		return ownerUserIds, "You have a grant request to approve (group receiver)", nil
	}

	managerIds, err := this.findDirectApprover(ctx, *grantRequest.ReceiverId, vErrs)
	fault.PanicOnErr(err)

	if len(managerIds) > 0 {
		return managerIds, "You have a grant request to approve", nil
	} else {
		return ownerUserIds, "You have a grant request to approve", nil
	}
}

func (this *GrantRequestServiceImpl) findDirectApprover(ctx crud.Context, userId model.Id, vErrs *fault.ValidationErrors) ([]string, error) {
	existCmd := &itUser.FindDirectApproverQuery{
		Id: userId,
	}
	existRes := itUser.FindDirectApproverResult{}

	err := this.cqrsBus.Request(ctx, *existCmd, &existRes)
	fault.PanicOnErr(err)

	if existRes.ClientError != nil {
		vErrs.MergeClientError(existRes.ClientError)
		return nil, nil
	}

	if len(existRes.Data) == 0 {
		return nil, nil
	}

	var managerIds []string
	for _, manager := range existRes.Data {
		if manager.Id != nil {
			managerIds = append(managerIds, string(*manager.Id))
		}
	}

	return managerIds, nil
}

func (this *GrantRequestServiceImpl) findOwner(ctx crud.Context, targetId string, targetType domain.GrantRequestTargetType, vErrs *fault.ValidationErrors) (*string, error) {
	switch targetType {
	case domain.GrantRequestTargetTypeRole:
		role, err := this.roleRepo.FindById(ctx, itRole.FindByIdParam{Id: targetId})
		fault.PanicOnErr(err)

		if role == nil {
			vErrs.AppendNotFound("targetRef", "target role")
			return nil, nil
		}
		return role.OwnerRef, nil

	case domain.GrantRequestTargetTypeSuite:
		suite, err := this.suiteRepo.FindById(ctx, itRoleSuite.FindByIdParam{Id: targetId})
		fault.PanicOnErr(err)

		if suite == nil {
			vErrs.AppendNotFound("targetRef", "target suite")
			return nil, nil
		}
		return suite.OwnerRef, nil

	default:
		return nil, nil
	}
}

func (this *GrantRequestServiceImpl) findOwnerInfo(ctx crud.Context, targetId string, targetType domain.GrantRequestTargetType, vErrs *fault.ValidationErrors) (*string, *string, error) {
	switch targetType {
	case domain.GrantRequestTargetTypeRole:
		role, err := this.roleRepo.FindById(ctx, itRole.FindByIdParam{Id: targetId})
		fault.PanicOnErr(err)

		if role == nil {
			vErrs.AppendNotFound("targetRef", "target role")
			return nil, nil, nil
		}
		ownerType := string(*role.OwnerType)
		return role.OwnerRef, &ownerType, nil

	case domain.GrantRequestTargetTypeSuite:
		suite, err := this.suiteRepo.FindById(ctx, itRoleSuite.FindByIdParam{Id: targetId})
		fault.PanicOnErr(err)

		if suite == nil {
			vErrs.AppendNotFound("targetRef", "target suite")
			return nil, nil, nil
		}
		ownerType := string(*suite.OwnerType)
		return suite.OwnerRef, &ownerType, nil

	default:
		return nil, nil, nil
	}
}

func (this *GrantRequestServiceImpl) findOwnerUserIds(ctx crud.Context, targetId string, targetType domain.GrantRequestTargetType, vErrs *fault.ValidationErrors) ([]string, error) {
	ownerId, ownerType, err := this.findOwnerInfo(ctx, targetId, targetType, vErrs)
	fault.PanicOnErr(err)

	if ownerId == nil {
		return nil, nil
	}

	if *ownerType == string(domain.RoleOwnerTypeUser) {
		return []string{*ownerId}, nil
	}

	if *ownerType == string(domain.RoleOwnerTypeGroup) {
		return this.getUsersInGroup(ctx, *ownerId)
	}

	return nil, nil
}

func (this *GrantRequestServiceImpl) getUsersInGroup(ctx crud.Context, groupId string) ([]string, error) {
	var allUserIds []string
	page := model.MODEL_RULE_PAGE_INDEX_START
	size := model.MODEL_RULE_PAGE_DEFAULT_SIZE

	for {
		graph := fmt.Sprintf("{\"if\":[\"groups.id\", \"=\", \"%s\"]}", groupId)
		searchParam := &crud.SearchQuery{
			Graph: &graph,
			Page:  &page,
			Size:  &size,
		}
		expandedUserQuery := &itUser.SearchUsersQuery{SearchQuery: *searchParam}

		searchRes := itUser.SearchUsersResult{}
		err := this.cqrsBus.Request(ctx, *expandedUserQuery, &searchRes)
		fault.PanicOnErr(err)

		if searchRes.ClientError != nil {
			return nil, searchRes.ClientError
		}

		if searchRes.Data == nil || searchRes.Data.Items == nil || len(searchRes.Data.Items) == 0 {
			break
		}

		for _, user := range searchRes.Data.Items {
			allUserIds = append(allUserIds, string(*user.Id))
		}

		currentPageCount := len(searchRes.Data.Items)
		totalFetched := len(allUserIds)

		if currentPageCount < size || totalFetched >= searchRes.Data.Total {
			break
		}

		page++
	}

	return allUserIds, nil
}

func (this *GrantRequestServiceImpl) assertGrantRequestExists(ctx crud.Context, grantRequest *domain.GrantRequest, vErrs *fault.ValidationErrors) (dbGrantRequest *domain.GrantRequest, err error) {
	dbGrantRequest, err = this.grantRequestRepo.FindById(ctx, itGrantRequest.FindByIdParam{Id: *grantRequest.Id})
	fault.PanicOnErr(err)

	if dbGrantRequest == nil {
		vErrs.AppendNotFound("grant_request_id", "grant request")
	}

	return
}

func (this *GrantRequestServiceImpl) assertCorrectEtag(updatedEtag model.Etag, dbEtag model.Etag, vErrs *fault.ValidationErrors) {
	if updatedEtag != dbEtag {
		vErrs.AppendEtagMismatched()
	}
}

func (this *GrantRequestServiceImpl) assertValidApprover(ctx crud.Context, request *domain.GrantRequest, responderId model.Id, vErrs *fault.ValidationErrors) ([]string, *string, error) {
	if request == nil {
		return nil, nil, nil
	}

	isValid, managerIds, ownerId, err := this.isValidApprover(ctx, request, responderId)
	fault.PanicOnErr(err)

	if !isValid {
		vErrs.Append("responder_id", "not authorized to approve this request")
		return nil, nil, nil
	}

	return managerIds, ownerId, nil
}

func (this *GrantRequestServiceImpl) isValidApprover(ctx crud.Context, request *domain.GrantRequest, responderId model.Id) (bool, []string, *string, error) {
	approvalCtx, err := this.buildApprovalContext(ctx, request, string(responderId))
	fault.PanicOnErr(err)

	if approvalCtx.HasAlreadyResponded() {
		return false, approvalCtx.ManagerIds, approvalCtx.OwnerId, nil
	}

	canApprove := this.determineCanApprove(approvalCtx)
	return canApprove, approvalCtx.ManagerIds, approvalCtx.OwnerId, nil
}

func (this *GrantRequestServiceImpl) buildApprovalContext(ctx crud.Context, request *domain.GrantRequest, responderId string) (*domain.ApprovalContext, error) {
	grantResponses, err := this.grantResponseRepo.FindByRequestId(ctx, *request.Id)
	fault.PanicOnErr(err)

	ownerId, err := this.findOwner(ctx, *request.TargetRef, *request.TargetType, nil)
	fault.PanicOnErr(err)

	ownerUserIds, err := this.findOwnerUserIds(ctx, *request.TargetRef, *request.TargetType, nil)
	fault.PanicOnErr(err)

	var managerIds []string
	if *request.ReceiverType == domain.ReceiverTypeUser {
		managerIds, err = this.findDirectApprover(ctx, *request.ReceiverId, nil)
		fault.PanicOnErr(err)
	}

	return &domain.ApprovalContext{
		Request:      request,
		ManagerIds:   managerIds,
		OwnerUserIds: ownerUserIds,
		OwnerId:      ownerId,
		ResponderId:  responderId,
		Responses:    grantResponses,
	}, nil
}

func (this *GrantRequestServiceImpl) determineCanApprove(approvalCtx *domain.ApprovalContext) bool {
	state := approvalCtx.GetResponseState()

	if approvalCtx.IsGroupReceiver() {
		if !approvalCtx.IsResponderOwnerUser() {
			return false
		}
		return !state.AnyOwnerDenied && !state.AnyOwnerApproved
	}

	if !approvalCtx.IsResponderManager() && !approvalCtx.IsResponderOwnerUser() {
		return false
	}

	if len(approvalCtx.ManagerIds) == 0 {
		return approvalCtx.IsResponderOwnerUser() && !state.AnyOwnerDenied && !state.AnyOwnerApproved
	}

	if state.AnyManagerDenied {
		return false
	}

	if !state.AnyManagerResponded {
		return approvalCtx.IsResponderManager()
	}

	if state.AnyManagerApproved {
		return approvalCtx.IsResponderOwnerUser() && !state.AnyOwnerDenied && !state.AnyOwnerApproved
	}

	return false
}

// Not implemented yet
func (this *GrantRequestServiceImpl) sendNotification(userId string, message string) error {
	return nil
}

func (this *GrantRequestServiceImpl) processGrantResponse(ctx crud.Context, dbGrantRequest *domain.GrantRequest, prevEtag model.Etag, cmd itGrantRequest.RespondToGrantRequestCommand, managerIds []string, ownerId *string) (*domain.GrantRequest, error) {
	err := this.createGrantResponse(ctx, cmd)
	fault.PanicOnErr(err)

	if cmd.Decision == domain.GrantRequestDecisionDeny {
		rejectedStatus := domain.RejectedGrantRequestStatus
		dbGrantRequest.Status = &rejectedStatus
		return this.handleGrantDenial(ctx, dbGrantRequest, prevEtag, cmd)
	}

	approvalCtx, err := this.buildApprovalContext(ctx, dbGrantRequest, string(cmd.ResponderId))
	fault.PanicOnErr(err)

	return this.handleGrantApproval(ctx, dbGrantRequest, prevEtag, cmd, approvalCtx)
}

func (this *GrantRequestServiceImpl) handleGrantApproval(ctx crud.Context, dbGrantRequest *domain.GrantRequest, prevEtag model.Etag, cmd itGrantRequest.RespondToGrantRequestCommand, approvalCtx *domain.ApprovalContext) (*domain.GrantRequest, error) {
	approvedStatus := domain.ApprovedGrantRequestStatus
	dbGrantRequest.Status = &approvedStatus

	if approvalCtx.IsGroupReceiver() || approvalCtx.IsResponderOwnerUser() {
		return this.handleFinalApproval(ctx, dbGrantRequest, prevEtag, cmd)
	}

	if approvalCtx.IsResponderManager() {
		return this.handleManagerApprovalWithGroupOwner(dbGrantRequest, approvalCtx.OwnerUserIds)
	}

	return dbGrantRequest, nil
}

func (this *GrantRequestServiceImpl) createGrantResponse(ctx crud.Context, cmd itGrantRequest.RespondToGrantRequestCommand) error {
	isApproved := cmd.Decision == domain.GrantRequestDecisionApprove

	grantResponse := &domain.GrantResponse{
		RequestId:   &cmd.Id,
		IsApproved:  &isApproved,
		Reason:      cmd.Reason,
		ResponderId: &cmd.ResponderId,
	}
	grantResponse.SetDefaults()

	_, err := this.grantResponseRepo.Create(ctx, *grantResponse)
	return err
}

func (this *GrantRequestServiceImpl) handleGrantDenial(ctx crud.Context, dbGrantRequest *domain.GrantRequest, prevEtag model.Etag, cmd itGrantRequest.RespondToGrantRequestCommand) (*domain.GrantRequest, error) {
	updatedGrantRequest, err := this.grantRequestRepo.Update(ctx, dbGrantRequest, prevEtag)
	fault.PanicOnErr(err)

	return updatedGrantRequest, this.sendNotification(*updatedGrantRequest.RequestorId, "Your grant request was rejected")
}

// func (this *GrantRequestServiceImpl) handleManagerApproval(dbGrantRequest *domain.GrantRequest, ownerId *string) (*domain.GrantRequest, error) {
// 	return dbGrantRequest, this.sendNotification(*ownerId, "Manager approved: You have a grant request to approve (final approval)")
// }

func (this *GrantRequestServiceImpl) handleManagerApprovalWithGroupOwner(dbGrantRequest *domain.GrantRequest, ownerUserIds []string) (*domain.GrantRequest, error) {
	for _, ownerUserId := range ownerUserIds {
		err := this.sendNotification(ownerUserId, "Manager approved: You have a grant request to approve (final approval)")
		fault.PanicOnErr(err)
	}
	return dbGrantRequest, nil
}

func (this *GrantRequestServiceImpl) handleFinalApproval(ctx crud.Context, dbGrantRequest *domain.GrantRequest, prevEtag model.Etag, cmd itGrantRequest.RespondToGrantRequestCommand) (*domain.GrantRequest, error) {
	dbGrantRequest.ApprovalId = &cmd.ResponderId

	err := this.grantAccess(ctx, dbGrantRequest)
	fault.PanicOnErr(err)

	// err = this.createPermissionHistory(ctx, dbGrantRequest, cmd)
	// fault.PanicOnErr(err)

	updatedGrantRequest, err := this.grantRequestRepo.Update(ctx, dbGrantRequest, prevEtag)
	fault.PanicOnErr(err)

	return updatedGrantRequest, this.sendNotification(*updatedGrantRequest.RequestorId, "Your grant request was approved and access has been granted")
}

func (this *GrantRequestServiceImpl) grantAccess(ctx crud.Context, request *domain.GrantRequest) error {
	switch *request.TargetType {
	case domain.GrantRequestTargetTypeRole:
		return this.createRoleUser(ctx, request)
	case domain.GrantRequestTargetTypeSuite:
		return this.createRoleSuiteUser(ctx, request)
	default:
		return &fault.ClientError{
			Code:    "bad_request",
			Details: "Invalid target type",
		}
	}
}

func (this *GrantRequestServiceImpl) createRoleUser(ctx crud.Context, request *domain.GrantRequest) error {
	err := this.roleRepo.AddRemoveUser(ctx, itRole.AddRemoveUserParam{
		Id:           *request.TargetRef,
		ApproverID:   *request.ApprovalId,
		ReceiverID:   *request.ReceiverId,
		ReceiverType: *request.ReceiverType,
		Add:          true,
	})
	fault.PanicOnErr(err)

	return nil
}

func (this *GrantRequestServiceImpl) createRoleSuiteUser(ctx crud.Context, request *domain.GrantRequest) error {
	err := this.suiteRepo.AddRemoveUser(ctx, itRoleSuite.AddRemoveUserParam{
		Id:           *request.TargetRef,
		ApproverID:   *request.ApprovalId,
		ReceiverID:   *request.ReceiverId,
		ReceiverType: *request.ReceiverType,
		Add:          true,
	})
	fault.PanicOnErr(err)

	return nil
}

// func (this *GrantRequestServiceImpl) createPermissionHistory(ctx crud.Context, dbGrantRequest *domain.GrantRequest, cmd itGrantRequest.RespondToGrantRequestCommand) error {
// 	receivers, err := this.getAffectedUsers(ctx, dbGrantRequest)
// 	fault.PanicOnErr(err)

// 	reason := this.determinePermissionHistoryReason(dbGrantRequest)

// 	for _, receiverId := range receivers {
// 		permissionHistory := this.buildPermissionHistory(dbGrantRequest, cmd, receiverId, reason)
// 		_, err := this.permissionHistoryRepo.Create(ctx, *permissionHistory)
// 		fault.PanicOnErr(err)
// 	}

// 	return nil
// }

// func (this *GrantRequestServiceImpl) getAffectedUsers(ctx crud.Context, dbGrantRequest *domain.GrantRequest) ([]model.Id, error) {
// 	if *dbGrantRequest.ReceiverType == domain.ReceiverTypeUser {
// 		return []model.Id{*dbGrantRequest.ReceiverId}, nil
// 	}

// 	receivers, err := this.getUsersInGroup(ctx, *dbGrantRequest.ReceiverId)
// 	fault.PanicOnErr(err)

// 	return receivers, nil
// }

// func (this *GrantRequestServiceImpl) determinePermissionHistoryReason(dbGrantRequest *domain.GrantRequest) domain.PermissionHistoryReason {
// 	isRole := *dbGrantRequest.TargetType == domain.GrantRequestTargetTypeRole
// 	isUser := *dbGrantRequest.ReceiverType == domain.ReceiverTypeUser

// 	if isRole {
// 		if isUser {
// 			return domain.PermissionHistoryReasonRoleAdded
// 		}
// 		return domain.PermissionHistoryReasonRoleAddedGroup
// 	} else {
// 		if isUser {
// 			return domain.PermissionHistoryReasonSuiteAdded
// 		}
// 		return domain.PermissionHistoryReasonSuiteAddedGroup
// 	}
// }

// func (this *GrantRequestServiceImpl) buildPermissionHistory(dbGrantRequest *domain.GrantRequest, cmd itGrantRequest.RespondToGrantRequestCommand, receiverId model.Id, reason domain.PermissionHistoryReason) *domain.PermissionHistory {
// 	effect := domain.PermissionHistoryEffectGrant

// 	permissionHistory := &domain.PermissionHistory{
// 		ApproverId:     &cmd.ResponderId,
// 		Effect:         &effect,
// 		Reason:         &reason,
// 		ReceiverId:     &receiverId,
// 		GrantRequestId: dbGrantRequest.Id,
// 	}

// 	permissionHistory.SetDefaults()

// 	switch *dbGrantRequest.TargetType {
// 	case domain.GrantRequestTargetTypeRole:
// 		permissionHistory.RoleId = dbGrantRequest.TargetRef
// 	case domain.GrantRequestTargetTypeSuite:
// 		permissionHistory.RoleSuiteId = dbGrantRequest.TargetRef
// 	}

//		return permissionHistory
//	}
func (this *GrantRequestServiceImpl) processDeleteGrantRequest(ctx crud.Context, dbGrantRequest *domain.GrantRequest) (int, error) {
	tx, err := this.grantRequestRepo.BeginTransaction(ctx)
	fault.PanicOnErr(err)

	ctx.SetDbTranx(tx)

	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "transaction process delete grant request"); e != nil {
			err = e
			tx.Rollback()
			return
		}

		tx.Commit()
	}()

	deletion, err := this.grantRequestRepo.Delete(ctx, itGrantRequest.DeleteParam{Id: *dbGrantRequest.Id})
	fault.PanicOnErr(err)

	if *dbGrantRequest.Status != domain.ApprovedGrantRequestStatus {
		return deletion, err
	}

	if *dbGrantRequest.TargetType == domain.GrantRequestTargetTypeRole {
		err := this.roleRepo.AddRemoveUser(ctx, itRole.AddRemoveUserParam{
			Id:           *dbGrantRequest.TargetRef,
			ReceiverID:   *dbGrantRequest.ReceiverId,
			ReceiverType: *dbGrantRequest.ReceiverType,
			Add:          false,
		})

		fault.PanicOnErr(err)
	}

	if *dbGrantRequest.TargetType == domain.GrantRequestTargetTypeSuite {
		err := this.suiteRepo.AddRemoveUser(ctx, itRoleSuite.AddRemoveUserParam{
			Id:           *dbGrantRequest.TargetRef,
			ReceiverID:   *dbGrantRequest.ReceiverId,
			ReceiverType: *dbGrantRequest.ReceiverType,
			Add:          false,
		})

		fault.PanicOnErr(err)
	}

	return deletion, err
}

func (this *GrantRequestServiceImpl) getGrantRequestById(ctx crud.Context, query itGrantRequest.GetGrantRequestByIdQuery, vErrs *fault.ValidationErrors) (dbGrantRequest *domain.GrantRequest, err error) {
	dbGrantRequest, err = this.grantRequestRepo.FindById(ctx, query)
	fault.PanicOnErr(err)

	if dbGrantRequest == nil {
		vErrs.AppendNotFound("id", "grant request id")
		return
	}

	err = this.populateGrantRequestDetails(ctx, dbGrantRequest)
	fault.PanicOnErr(err)

	return
}

func (this *GrantRequestServiceImpl) populateGrantRequestDetails(ctx crud.Context, dbGrantRequest *domain.GrantRequest) (err error) {
	dbGrantRequest.RequestorName, err = this.getUserDisplayName(ctx, *dbGrantRequest.RequestorId, "user", nil)
	fault.PanicOnErr(err)

	if *dbGrantRequest.ReceiverType == domain.ReceiverTypeGroup {
		dbGrantRequest.ReceiverName, err = this.getUserDisplayName(ctx, *dbGrantRequest.ReceiverId, "group", nil)
	} else {
		dbGrantRequest.ReceiverName, err = this.getUserDisplayName(ctx, *dbGrantRequest.ReceiverId, "user", nil)
	}
	fault.PanicOnErr(err)

	for i, grantResponse := range dbGrantRequest.GrantResponses {
		dbGrantRequest.GrantResponses[i].ResponderName, err = this.getUserDisplayName(ctx, *grantResponse.ResponderId, "user", nil)
		fault.PanicOnErr(err)
	}

	// Populate organization
	if dbGrantRequest.OrgId != nil {
		org, err := this.getOrganizationById(ctx, *dbGrantRequest.OrgId)
		fault.PanicOnErr(err)
		if org != nil {
			dbGrantRequest.Organization = org
			dbGrantRequest.OrgName = org.DisplayName
		}
	}

	return
}

func (this *GrantRequestServiceImpl) getUserDisplayName(ctx crud.Context, id model.Id, entityType string, vErrs *fault.ValidationErrors) (*string, error) {
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

func (this *GrantRequestServiceImpl) getOrganizationById(ctx crud.Context, orgId model.Id) (*domain.Organization, error) {
	orgQuery := &itOrg.GetOrganizationByIdQuery{
		Id: orgId,
	}
	orgRes := itOrg.GetOrganizationByIdResult{}
	err := this.cqrsBus.Request(ctx, *orgQuery, &orgRes)
	fault.PanicOnErr(err)

	if orgRes.ClientError != nil {
		return nil, nil
	}

	if orgRes.Data == nil {
		return nil, nil
	}

	return &domain.Organization{
		Id:          orgRes.Data.Id,
		DisplayName: orgRes.Data.DisplayName,
	}, nil
}

func (this *GrantRequestServiceImpl) findGrantRequestsByTarget(ctx crud.Context, targetType domain.GrantRequestTargetType, targetRef model.Id) ([]domain.GrantRequest, error) {
	return this.grantRequestRepo.FindAllByTarget(ctx, itGrantRequest.FindAllByTargetParam{
		TargetType: targetType,
		TargetRef:  targetRef,
	})
}
