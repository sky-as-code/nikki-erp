// Package auditlog records who did what to which record: every mutation the composable engine
// performs, the reads a resource opts into, and the custom actions that report themselves.
//
// Writers reach it through the core/audit.AuditSink port, which this module implements and
// apptrait registers; they never import this module. It must therefore never depend on a feature
// module, since every one of them writes through it.
package auditlog

import (
	stdErr "errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/auditlog/app"
	c "github.com/sky-as-code/nikki-erp/modules/auditlog/constants"
	"github.com/sky-as-code/nikki-erp/modules/auditlog/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/auditlog/dynamicengines"
	"github.com/sky-as-code/nikki-erp/modules/auditlog/transport"
)

// ModuleSingleton is typed DynamicModule so that dropping RegisterModels fails the build.
var ModuleSingleton modules.DynamicModule = &AuditLogModule{}

type AuditLogModule struct {
}

func (*AuditLogModule) LabelKey() string {
	return "auditlog.moduleLabel"
}

func (*AuditLogModule) Name() string {
	return c.AuditLogModuleName
}

func (*AuditLogModule) ModelPrefixes() []string {
	return []string{c.AuditLogModuleName}
}

func (*AuditLogModule) HasUi() bool {
	return false
}

func (*AuditLogModule) Category() modules.ModuleCategory {
	return modules.ModuleCategoryFoundation
}

// Deps implements InCodeModule. iam is absent on purpose: the actor_user edge declares no
// constraint, and iam writes its own records through this module's port.
func (*AuditLogModule) Deps() []string {
	return []string{"dynamicresource"}
}

func (*AuditLogModule) IsInternal() bool {
	return false
}

func (*AuditLogModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init builds the engine before anything binds to it, and transport last.
func (*AuditLogModule) Init() error {
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}
	return stdErr.Join(
		app.InitApplicationServices(),
		transport.InitTransport(),
	)
}

func (*AuditLogModule) RegisterModels() error {
	return dmodel.RegisterSchemaB(models.ActionSchemaBuilder())
}
