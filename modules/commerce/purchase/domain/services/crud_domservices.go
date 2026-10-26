package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
	itAgreement "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/agreement"
	itAudit "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/audit"
	itConfiguration "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/configuration"
	itOrder "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/order"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The CRUD domain services of the resources that add no write rule of their own, and of the two a
// client may not write. A refused write stays a defined action that answers a business violation,
// so the caller gets a 400 naming the reason instead of a 404 that reads as a wrong URL.

func NewConfigurationCrudDomainService(base composable.CrudDomainService) itConfiguration.ConfigurationCrudDomainService {
	return &ConfigurationCrudDomainServiceImpl{CrudDomainService: base}
}

type ConfigurationCrudDomainServiceImpl struct {
	composable.CrudDomainService
}

func NewAgreementLineCrudDomainService(base composable.CrudDomainService) itAgreement.AgreementLineCrudDomainService {
	return &AgreementLineCrudDomainServiceImpl{CrudDomainService: base}
}

type AgreementLineCrudDomainServiceImpl struct {
	composable.CrudDomainService
}

// NewAuditEventCrudDomainService closes the audit trail to clients. An event is written only by
// the audit helper, through the repository and inside the transaction of the transition it
// records; a client-written one would be indistinguishable from a real event. Update and delete are
// refused because the record is immutable.
func NewAuditEventCrudDomainService(base composable.CrudDomainService) itAudit.AuditEventCrudDomainService {
	return &AuditEventCrudDomainServiceImpl{CrudDomainService: base}
}

type AuditEventCrudDomainServiceImpl struct {
	composable.CrudDomainService
}

func (this *AuditEventCrudDomainServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand, options ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(rejectAuditEventWrite, opts.ValidateExtra)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *AuditEventCrudDomainServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	opts := safe.GetOptional(options, composable.UpdateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(rejectAuditEventWrite, opts.ValidateExtra)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

func (this *AuditEventCrudDomainServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	opts := safe.GetOptional(options, composable.DeleteOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(rejectAuditEventWrite, opts.ValidateExtra)
	return this.CrudDomainService.Delete(ctx, cmd, opts)
}

// rejectAuditEventWrite is the refusal every client write of an audit event answers with.
func rejectAuditEventWrite(
	_ corectx.Context, _ *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	vErrs.Append(*ft.NewBusinessViolation(
		models.AuditEventSchemaName,
		ft.ErrorKey("purchase_audit_event.not_client_writable", "purchase"),
		"audit events are written by the system when a purchase order or agreement changes state; "+
			"they cannot be created, edited or removed",
	))
	return nil
}

// NewSourcingGroupCrudDomainService closes direct creation and deletion of a sourcing group. The
// group is created by create_alternative and reaped when it drops to one order or fewer, both
// through the repository; a hand-made one would be an empty container nothing reaps and a
// hand-deleted one would strand its orders. Update stays open.
func NewSourcingGroupCrudDomainService(base composable.CrudDomainService) itOrder.SourcingGroupCrudDomainService {
	return &SourcingGroupCrudDomainServiceImpl{CrudDomainService: base}
}

type SourcingGroupCrudDomainServiceImpl struct {
	composable.CrudDomainService
}

func (this *SourcingGroupCrudDomainServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand, options ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(rejectSourcingGroupWrite, opts.ValidateExtra)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *SourcingGroupCrudDomainServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	opts := safe.GetOptional(options, composable.DeleteOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(rejectSourcingGroupWrite, opts.ValidateExtra)
	return this.CrudDomainService.Delete(ctx, cmd, opts)
}

func rejectSourcingGroupWrite(
	_ corectx.Context, _ *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	vErrs.Append(*ft.NewBusinessViolation(
		models.SourcingGroupSchemaName,
		ft.ErrorKey("purchase_sourcing_group.not_client_writable", "purchase"),
		"sourcing groups are created by adding an alternative to a purchase order and removed when "+
			"fewer than two alternatives remain",
	))
	return nil
}
