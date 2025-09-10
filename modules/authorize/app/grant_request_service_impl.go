package app

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_request"
	itGrantResponse "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_response"
	itPermissionHistory "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/permission_history"
	itRole "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
	itRoleSuite "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
	itGroup "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type ApprovalType int

const (
	ApprovalTypeManagerOnly ApprovalType = iota
	ApprovalTypeManagerAndOwner
	ApprovalTypeOwnerOnly
	ApprovalTypeNone
)

func NewGrantRequestServiceImpl(grantRequestRepo itGrantRequest.GrantRequestRepository, grantResponseRepo itGrantResponse.GrantResponseRepository, roleRepo itRole.RoleRepository, suiteRepo itRoleSuite.RoleSuiteRepository, permissionHistoryRepo itPermissionHistory.PermissionHistoryRepository, eventBus event.EventBus, cqrsBus cqrs.CqrsBus) itGrantRequest.GrantRequestService {
	return &GrantRequestServiceImpl{
		grantRequestRepo:      grantRequestRepo,
		grantResponseRepo:     grantResponseRepo,
		roleRepo:              roleRepo,
		suiteRepo:             suiteRepo,
		permissionHistoryRepo: permissionHistoryRepo,
		eventBus:              eventBus,
		cqrsBus:               cqrsBus,
	}
}

type GrantRequestServiceImpl struct {
	grantRequestRepo      itGrantRequest.GrantRequestRepository
	grantResponseRepo     itGrantResponse.GrantResponseRepository
	roleRepo              itRole.RoleRepository
	suiteRepo             itRoleSuite.RoleSuiteRepository
	permissionHistoryRepo itPermissionHistory.PermissionHistoryRepository
	eventBus              event.EventBus
	cqrsBus               cqrs.CqrsBus
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
			return this.setupApprovalChain(ctx, grantRequest, vErrs)
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itGrantRequest.CreateGrantRequestResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	createdGrantRequest, err := this.grantRequestRepo.Create(ctx, *grantRequest)
	fault.PanicOnErr(err)

	return &itGrantRequest.CreateGrantRequestResult{
		Data:    createdGrantRequest,
		HasData: createdGrantRequest != nil,
	}, nil
}

func (this *GrantRequestServiceImpl) RespondToGrantRequest(ctx crud.Context, cmd itGrantRequest.RespondToGrantRequestCommand) (result *itGrantRequest.RespondToGrantRequestResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "respond to grant request"); e != nil {
			err = e
		}
	}()

	var dbGrantRequest *domain.GrantRequest

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbGrantRequest, err = this.assertGrantRequestExists(ctx, cmd.Id, vErrs)
			return err
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			return this.assertValidApprover(ctx, dbGrantRequest, cmd.ResponderId, vErrs)
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itGrantRequest.RespondToGrantRequestResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	respondGrantRequest, err := this.processGrantResponse(ctx, dbGrantRequest, cmd)
	fault.PanicOnErr(err)

	return &itGrantRequest.RespondToGrantRequestResult{
		Data:    respondGrantRequest,
		HasData: true,
	}, nil
}

func (this *GrantRequestServiceImpl) setGrantRequestDefaults(grantRequest *domain.GrantRequest) {
	grantRequest.SetDefaults()
}

func (this *GrantRequestServiceImpl) assertTarget(ctx crud.Context, grantRequest *domain.GrantRequest, vErrs *fault.ValidationErrors) {
	switch *grantRequest.TargetType {
	case domain.GrantRequestTargetTypeRole:
		role, err := this.roleRepo.FindById(ctx, itRole.FindByIdParam{Id: *grantRequest.TargetRef})
		fault.PanicOnErr(err)

		if role == nil {
			vErrs.AppendNotFound("targetRef", "target")
			return
		}
		this.validateTarget(role.IsRequestable, role.IsRequiredAttachment, role.IsRequiredComment, grantRequest, vErrs)
	case domain.GrantRequestTargetTypeSuite:
		suite, err := this.suiteRepo.FindById(ctx, itRoleSuite.FindByIdParam{Id: *grantRequest.TargetRef})
		fault.PanicOnErr(err)

		if suite == nil {
			vErrs.AppendNotFound("targetRef", "target")
			return
		}
		this.validateTarget(suite.IsRequestable, suite.IsRequiredAttachment, suite.IsRequiredComment, grantRequest, vErrs)
	}
}

func (this *GrantRequestServiceImpl) assertReceiver(ctx crud.Context, grantRequest *domain.GrantRequest, vErrs *fault.ValidationErrors) {
	switch *grantRequest.ReceiverType {
	case domain.ReceiverTypeUser:
		existCmd := &itUser.UserExistsCommand{
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
			vErrs.Append("receiver_id", "not existing user")
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
			vErrs.Append("receiver_id", "not existing group")
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
		if grantRequest.AttachmentUrl == nil || *grantRequest.AttachmentUrl == "" {
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
		})
		fault.PanicOnErr(err)

		if exist {
			vErrs.AppendAlreadyExists("receiver_id", "receiver")
		}
	case domain.GrantRequestTargetTypeSuite:
		exist, err := this.suiteRepo.ExistUserWithRoleSuite(ctx, itRoleSuite.ExistUserWithRoleSuiteParam{
			ReceiverType: *grantRequest.ReceiverType,
			ReceiverId:   *grantRequest.ReceiverId,
		})
		fault.PanicOnErr(err)

		if exist {
			vErrs.AppendAlreadyExists("receiver_id", "receiver")
		}
	}
}

func (this *GrantRequestServiceImpl) assertNoPendingGrantRequest(ctx crud.Context, cmd itGrantRequest.CreateGrantRequestCommand, vErrs *fault.ValidationErrors) error {
	pendingRequests, err := this.grantRequestRepo.FindPendingByReceiverAndTarget(ctx, cmd.ReceiverId, cmd.TargetRef, domain.GrantRequestTargetType(cmd.TargetType))
	fault.PanicOnErr(err)

	if len(pendingRequests) > 0 {
		vErrs.AppendAlreadyExists("receiver_id", "receiver already has a pending request for this role/suite")
	}

	return nil
}

func (this *GrantRequestServiceImpl) setupApprovalChain(ctx crud.Context, grantRequest *domain.GrantRequest, vErrs *fault.ValidationErrors) error {
	status := domain.PendingGrantRequestStatus
	grantRequest.Status = &status

	managerId, err := this.findDirectApprover(ctx, *grantRequest.ReceiverId, vErrs)
	fault.PanicOnErr(err)

	if managerId != nil {
		err = this.sendNotification(*managerId, "You have a grant request to approve")
		fault.PanicOnErr(err)

		grantRequest.ApprovalId = managerId
	} else {
		ownerId, err := this.findOwner(ctx, *grantRequest.TargetRef, *grantRequest.TargetType, vErrs)
		fault.PanicOnErr(err)

		err = this.sendNotification(*ownerId, "You have a grant request to approve")
		fault.PanicOnErr(err)

		grantRequest.ApprovalId = ownerId
	}

	return nil
}

func (this *GrantRequestServiceImpl) findDirectApprover(ctx crud.Context, userId model.Id, vErrs *fault.ValidationErrors) (*string, error) {
	existCmd := &itUser.FindDirectApproverQuery{
		Id: userId,
	}
	existRes := itUser.FindDirectApproverResult{}

	err := this.cqrsBus.Request(ctx, *existCmd, &existRes)
	if err != nil {
		return nil, err
	}

	if existRes.ClientError != nil {
		vErrs.MergeClientError(existRes.ClientError)
		return nil, nil
	}

	if existRes.Data == nil {
		return nil, nil
	}

	return existRes.Data.Id, nil
}

func (this *GrantRequestServiceImpl) findOwner(ctx crud.Context, targetId string, targetType domain.GrantRequestTargetType, vErrs *fault.ValidationErrors) (*string, error) {
	switch targetType {
	case domain.GrantRequestTargetTypeRole:
		role, err := this.roleRepo.FindById(ctx, itRole.FindByIdParam{Id: targetId})
		fault.PanicOnErr(err)

		if role == nil {
			vErrs.AppendNotFound("targetRef", "target")
			return nil, nil
		}
		return role.OwnerRef, nil

	case domain.GrantRequestTargetTypeSuite:
		suite, err := this.suiteRepo.FindById(ctx, itRoleSuite.FindByIdParam{Id: targetId})
		fault.PanicOnErr(err)

		if suite == nil {
			vErrs.AppendNotFound("targetRef", "target")
			return nil, nil
		}
		return suite.OwnerRef, nil

	default:
		return nil, nil
	}
}

// Not implemented yet
func (this *GrantRequestServiceImpl) sendNotification(userId string, message string) error {
	return nil
}

func (this *GrantRequestServiceImpl) processGrantResponse(ctx crud.Context, dbGrantRequest *domain.GrantRequest, cmd itGrantRequest.RespondToGrantRequestCommand) (*domain.GrantRequest, error) {
	if cmd.Decision == domain.GrantRequestDecisionDeny {
		return this.handleGrantDenial(ctx, dbGrantRequest)
	}

	err := this.createGrantResponse(ctx, cmd)
	fault.PanicOnErr(err)

	managerId, ownerId, err := this.getApprovalChainInfo(ctx, dbGrantRequest)
	fault.PanicOnErr(err)

	approvalType := this.determineApprovalType(cmd.ResponderId, managerId, ownerId)

	switch approvalType {
	case ApprovalTypeManagerOnly:
		return this.handleManagerApproval(ctx, dbGrantRequest, ownerId)
	case ApprovalTypeManagerAndOwner:
		return this.handleFinalApproval(ctx, dbGrantRequest, cmd.ResponderId)
	case ApprovalTypeOwnerOnly:
		return this.handleFinalApproval(ctx, dbGrantRequest, cmd.ResponderId)
	default:
		return dbGrantRequest, nil
	}
}

func (this *GrantRequestServiceImpl) handleGrantDenial(ctx crud.Context, dbGrantRequest *domain.GrantRequest) (*domain.GrantRequest, error) {
	status := domain.RejectedGrantRequestStatus
	dbGrantRequest.Status = &status

	updatedGrantRequest, err := this.grantRequestRepo.Update(ctx, *dbGrantRequest)
	fault.PanicOnErr(err)

	return updatedGrantRequest, this.sendNotification(*updatedGrantRequest.RequestorId, "Your grant request was rejected")
}

func (this *GrantRequestServiceImpl) createGrantResponse(ctx crud.Context, cmd itGrantRequest.RespondToGrantRequestCommand) error {
	isApproved := true
	grantResponse := &domain.GrantResponse{
		RequestId:   &cmd.Id,
		IsApproved:  &isApproved,
		Reason:      nil,
		ResponderId: &cmd.ResponderId,
	}
	grantResponse.SetDefaults()

	_, err := this.grantResponseRepo.Create(ctx, *grantResponse)
	return err
}

func (this *GrantRequestServiceImpl) getApprovalChainInfo(ctx crud.Context, dbGrantRequest *domain.GrantRequest) (*string, *string, error) {
	managerId, err := this.findDirectApprover(ctx, *dbGrantRequest.ReceiverId, nil)
	if err != nil {
		return nil, nil, err
	}

	ownerId, err := this.findOwner(ctx, *dbGrantRequest.TargetRef, *dbGrantRequest.TargetType, nil)
	if err != nil {
		return nil, nil, err
	}

	return managerId, ownerId, nil
}

func (this *GrantRequestServiceImpl) determineApprovalType(responderId model.Id, managerId, ownerId *string) ApprovalType {
	responderIdStr := string(responderId)

	isManager := managerId != nil && responderIdStr == *managerId
	isOwner := ownerId != nil && responderIdStr == *ownerId
	isManagerAndOwner := isManager && isOwner

	if isManagerAndOwner {
		return ApprovalTypeManagerAndOwner
	} else if isManager {
		return ApprovalTypeManagerOnly
	} else if isOwner {
		return ApprovalTypeOwnerOnly
	}

	return ApprovalTypeNone
}

func (this *GrantRequestServiceImpl) handleManagerApproval(ctx crud.Context, dbGrantRequest *domain.GrantRequest, ownerId *string) (*domain.GrantRequest, error) {
	return dbGrantRequest, this.sendNotification(*ownerId, "Manager approved: You have a grant request to approve (final approval)")
}

func (this *GrantRequestServiceImpl) handleFinalApproval(ctx crud.Context, dbGrantRequest *domain.GrantRequest, approverId model.Id) (*domain.GrantRequest, error) {
	status := domain.ApprovedGrantRequestStatus
	dbGrantRequest.Status = &status
	approverIdStr := string(approverId)
	dbGrantRequest.ApprovalId = &approverIdStr

	err := this.grantAccess(ctx, dbGrantRequest)
	fault.PanicOnErr(err)

	err = this.createPermissionHistory(ctx, dbGrantRequest, approverId)
	fault.PanicOnErr(err)

	updatedGrantRequest, err := this.grantRequestRepo.Update(ctx, *dbGrantRequest)
	fault.PanicOnErr(err)

	return updatedGrantRequest, this.sendNotification(*updatedGrantRequest.RequestorId, "Your grant request was approved and access has been granted")
}

func (this *GrantRequestServiceImpl) assertGrantRequestExists(ctx crud.Context, grantRequestID model.Id, vErrs *fault.ValidationErrors) (dbGrantRequest *domain.GrantRequest, err error) {
	dbGrantRequest, err = this.grantRequestRepo.FindById(ctx, itGrantRequest.FindByIdParam{Id: grantRequestID})
	fault.PanicOnErr(err)

	if dbGrantRequest == nil {
		vErrs.AppendNotFound("grant_request_id", "grant request")
	} else if *dbGrantRequest.Status != domain.PendingGrantRequestStatus {
		vErrs.Append("grant_request_id", "grant request is not pending")

	}

	return
}

func (this *GrantRequestServiceImpl) assertValidApprover(ctx crud.Context, request *domain.GrantRequest, responderId model.Id, vErrs *fault.ValidationErrors) error {
	if request == nil {
		return nil
	}

	isValid, err := this.isValidApprover(ctx, request, responderId)
	fault.PanicOnErr(err)

	if !isValid {
		vErrs.Append("responder_id", "not authorized to approve this request")
	}

	return nil
}

func (this *GrantRequestServiceImpl) isValidApprover(ctx crud.Context, request *domain.GrantRequest, responderId model.Id) (bool, error) {
	managerId, err := this.findDirectApprover(ctx, *request.ReceiverId, nil)
	if err != nil {
		return false, err
	}

	ownerId, err := this.findOwner(ctx, *request.TargetRef, *request.TargetType, nil)
	if err != nil {
		return false, err
	}

	if managerId != nil && string(responderId) == *managerId {
		return true, nil // Manager can approve
	}

	if ownerId != nil && string(responderId) == *ownerId {
		return true, nil // Owner can approve
	}

	return false, nil
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

func (this *GrantRequestServiceImpl) createPermissionHistory(ctx crud.Context, request *domain.GrantRequest, approverId model.Id) error {
	// TODO: Implement Permission History creation
	// This will require PermissionHistory repository
	// reason := "role_added"
	// if *request.TargetType == "suite" {
	//     reason = "suite_added"
	// }
	//
	// permissionHistory := &domain.PermissionHistory{
	//     Effect: "grant",
	//     Reason: reason,
	//     ReceiverId: request.ReceiverId,
	//     GrantRequestId: request.Id,
	//     ApproverId: &approverId,
	//     // ... other fields
	// }
	// return this.permissionHistoryRepo.Create(ctx, *permissionHistory)
	return nil
}
