package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable/storagefield"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

// NewUserCrudDomainService is handed the composable default by the iam_user onion.
func NewUserCrudDomainService(
	base composable.CrudDomainService,
	settingsSvc itExt.UserSettingsExtService,
	permRepo itPerm.PermissionRepository,
	historyRepo itPerm.PermissionHistoryRepository,
	fileSvc itExt.FileStorageExtService,
) itUser.UserCrudDomainService {
	return &UserCrudDomainServiceImpl{
		CrudDomainService: base,
		settingsSvc:       settingsSvc,
		permRepo:          permRepo,
		auditor:           permissionAuditor{historyRepo: historyRepo},
		// A user belongs to the tenant, not to an org, so its avatar may be any tenant file.
		files: storagefield.NewGuard(fileSvc, storagefield.ScopeTenant, domain.UserFieldAvatarFilePath),
	}
}

type UserCrudDomainServiceImpl struct {
	composable.CrudDomainService
	settingsSvc itExt.UserSettingsExtService
	permRepo    itPerm.PermissionRepository
	auditor     permissionAuditor
	files       storagefield.Guard
}

// ManageRoles assigns roles to and removes them from the user, then refreshes the user's
// permission cache so authorization reflects the change immediately. The trail and the grant
// share the transaction: they stand or fall together.
func (this *UserCrudDomainServiceImpl) ManageRoles(
	ctx corectx.Context, cmd itUser.ManageUserRolesCommand,
) (*itUser.ManageUserRolesResult, error) {
	return composable.ManageM2m(ctx, this.Repository(), cmd, composable.ManageM2mParam{
		Action:          "manage user role assignments",
		DestSchemaName:  domain.RoleSchemaName,
		JunctionIdField: domain.RoleUserAssignFieldId,
		AfterChange: func(tranxCtx corectx.Context, change composable.M2mChange) error {
			if err := this.permRepo.RebuildUserPermission(tranxCtx, change.SrcId); err != nil {
				return err
			}
			return this.auditor.recordRoleChange(tranxCtx, change.Added, change.Removed,
				[]model.Id{change.SrcId},
				domain.PermissionHistoryReasonRoleAdded, domain.PermissionHistoryReasonRoleRemoved)
		},
	})
}

// Create creates the user, seeds their preferences and builds their permission cache in one
// transaction.
//
// The seeding shares the transaction so that a user never exists without the preference rows the
// application expects them to have: their first sign-in reads a theme and a language, and there is
// no later point that would repair a user created without them. A new user holds no assignment
// yet, but does hold every system role: without the rebuild their cache stays empty, and the
// inbox, among others, refuses them until something else happens to rebuild it.
func (this *UserCrudDomainServiceImpl) Create(
	ctx corectx.Context, cmd itUser.CreateUserCrudCommand, options ...composable.CreateOptions,
) (*itUser.CreateUserCrudResult, error) {
	opts := composable.CreateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.BeforeValidation = clearIsOwner(opts.BeforeValidation)
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.files.ValidateCreate())

	var result *itUser.CreateUserCrudResult
	err := composable.InTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		var err error
		result, err = this.CrudDomainService.Create(tranxCtx, cmd, opts)
		// A rejected create has written nothing, so there is nothing to seed.
		if err != nil || result.ClientErrors.Count() > 0 || !result.HasData {
			return err
		}
		return this.afterUserCreated(tranxCtx, result.Data)
	})
	if err != nil {
		return nil, errors.Wrap(err, "create user")
	}
	return result, nil
}

// Update checks a changed avatar_file_path against IT Storage and removes the replaced file.
func (this *UserCrudDomainServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	opts := composable.UpdateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	return this.files.Update(ctx, opts, func(opts composable.UpdateOptions) (*composable.MutateResult, error) {
		return this.CrudDomainService.Update(ctx, cmd, opts)
	})
}

// clearIsOwner drops is_owner from a create: the owner is set up by the installer, and a normal
// user must never carry the flag.
func clearIsOwner(next composable.BeforeValidationFn) composable.BeforeValidationFn {
	return func(
		ctx corectx.Context, entity *composable.DynamicEntity, vErrs *ft.ClientErrors,
	) (*composable.DynamicEntity, error) {
		delete(entity.GetFieldData(), domain.UserFieldIsOwner)
		if next == nil {
			return entity, nil
		}
		return next(ctx, entity, vErrs)
	}
}

func (this *UserCrudDomainServiceImpl) afterUserCreated(ctx corectx.Context, created dmodel.DynamicFields) error {
	userId := composable.ParamId(created)
	if userId == "" {
		return errors.New("the created user has no id")
	}
	initResult, err := this.settingsSvc.InitUserPreferences(ctx, itExt.InitOwnerSettingsCommand{OwnerId: userId})
	if err != nil {
		return err
	}
	// A rejected seeding is a defect in this call rather than something the caller creating a user
	// can correct, so it fails the whole create.
	if initResult.ClientErrors.Count() > 0 {
		return initResult.ClientErrors.ToError()
	}
	return this.permRepo.RebuildUserPermission(ctx, userId)
}
