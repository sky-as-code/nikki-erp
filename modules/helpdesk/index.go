package helpdesk

import (
	"errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/app"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
	repo "github.com/sky-as-code/nikki-erp/modules/helpdesk/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/transport"
)

var ModuleSingleton modules.InCodeModule = &HelpdeskModule{}

type HelpdeskModule struct{}

func (*HelpdeskModule) LabelKey() string { return "helpdesk.moduleLabel" }
func (*HelpdeskModule) Name() string     { return "helpdesk" }
func (*HelpdeskModule) Deps() []string   { return []string{} }
func (*HelpdeskModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

func (*HelpdeskModule) Init() error {
	return errors.Join(
		repo.InitRepositories(),
		app.InitServices(),
		transport.InitTransport(),
	)
}

func (*HelpdeskModule) RegisterModels() error {
	return errors.Join(
		dmodel.RegisterSchemaB(domain.TicketCategoryRelSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.SlaPolicySchemaBuilder()),
		dmodel.RegisterSchemaB(domain.TeamSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.TicketCategorySchemaBuilder()),
		dmodel.RegisterSchemaB(domain.TicketSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.TicketActivitySchemaBuilder()),
		dmodel.RegisterSchemaB(domain.TicketMessageSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.TicketAssignmentSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.SlaBreachSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.TeamMembershipSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.EscalationRuleSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.TicketFeedbackSchemaBuilder()),
	)
}
