package helpdesk

import (
	"errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/app"
	modconstants "github.com/sky-as-code/nikki-erp/modules/helpdesk/constants"
	models "github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/helpdesk/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/transport"
)

var ModuleSingleton modules.InCodeModule = &HelpdeskModule{}

type HelpdeskModule struct{}

func (*HelpdeskModule) LabelKey() string { return "helpdesk.moduleLabel" }
func (*HelpdeskModule) Name() string     { return modconstants.HelpdeskModuleName }
func (*HelpdeskModule) Deps() []string   { return []string{} }
func (*HelpdeskModule) IsInternal() bool { return false }
func (*HelpdeskModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

func (*HelpdeskModule) Init() error {
	return errors.Join(
		repo.InitRepositories(),
		services.InitDomainServices(),
		app.InitApplicationServices(),
		transport.InitTransport(),
	)
}

func (*HelpdeskModule) RegisterModels() error {
	return errors.Join(
		dmodel.RegisterSchemaB(models.TicketCategoryRelSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SlaPolicySchemaBuilder()),
		dmodel.RegisterSchemaB(models.TeamSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TicketCategorySchemaBuilder()),
		dmodel.RegisterSchemaB(models.TicketSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TicketActivitySchemaBuilder()),
		dmodel.RegisterSchemaB(models.TicketMessageSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TicketAssignmentSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SlaBreachSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TeamMembershipSchemaBuilder()),
		dmodel.RegisterSchemaB(models.EscalationRuleSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TicketFeedbackSchemaBuilder()),
	)
}
