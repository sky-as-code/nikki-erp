package app

import (
	"strings"

	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itRevokeRequest "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/revoke_request"
	itRole "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
	itSuite "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
)

func NewRevokeRequestServiceImpl(revokeRequestRepo itRevokeRequest.RevokeRequestRepository, roleRepo itRole.RoleRepository, suiteRepo itSuite.RoleSuiteRepository) itRevokeRequest.RevokeRequestService {
	return &RevokeRequestServiceImpl{
		revokeRequestRepo: revokeRequestRepo,
		roleRepo:          roleRepo,
		suiteRepo:         suiteRepo,
	}
}

type RevokeRequestServiceImpl struct {
	revokeRequestRepo itRevokeRequest.RevokeRequestRepository
	roleRepo          itRole.RoleRepository
	suiteRepo         itSuite.RoleSuiteRepository
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

	if revokeRequest.AttachmentUrl != nil {
		cleanedAttachmentUrl := strings.TrimSpace(*revokeRequest.AttachmentUrl)
		cleanedAttachmentUrl = defense.SanitizePlainText(cleanedAttachmentUrl)
		revokeRequest.AttachmentUrl = &cleanedAttachmentUrl
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

func (this *RevokeRequestServiceImpl) classifyReceiverType(targetType *domain.ReceiverType) domain.ReceiverType {
	if *targetType == domain.ReceiverTypeUser {
		return domain.ReceiverTypeUser
	} else {
		return domain.ReceiverTypeGroup
	}
}
