package core

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/authtoken"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	coredyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/enum"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient"
	http "github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/i18n"
	"github.com/sky-as-code/nikki-erp/modules/core/infra"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/core/tag"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.InCodeModule = &CoreModule{}

type CoreModule struct {
}

// Name implements NikkiModule.
func (*CoreModule) Name() string {
	return "core"
}

// LabelKey implements NikkiModule.
func (*CoreModule) LabelKey() string {
	return "core.moduleLabel"
}

// Deps implements NikkiModule.
func (*CoreModule) Deps() []string {
	return []string{
		"apptrait",
	}
}

// Version implements NikkiModule.
func (*CoreModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements NikkiModule.
func (*CoreModule) Init() error {
	err := errors.Join(
		deps.Invoke(config.InitSubModule),
		deps.Invoke(cqrs.InitSubModule),
		deps.Invoke(authtoken.InitSubModule),
		deps.Invoke(event.InitSubModule),
		deps.Invoke(db.InitSubModule),
		deps.Invoke(coredyn.InitSubModule),
		deps.Invoke(http.InitSubModule),
		deps.Invoke(httpclient.InitSubModule),
		deps.Register(newCoreClient),
		deps.Invoke(enum.InitSubModule),
		deps.Invoke(tag.InitSubModule),
		deps.Invoke(infra.InitSubModule),
		// These submodules expose network APIs
		deps.Invoke(i18n.InitSubModule),
	)

	return err
}

func newCoreClient(clientOpts *db.EntClientOptions) *ent.Client {
	var client *ent.Client
	if clientOpts.DebugEnabled {
		client = ent.NewClient(ent.Driver(clientOpts.Driver), ent.Debug())
	}
	client = ent.NewClient(ent.Driver(clientOpts.Driver))

	err := client.DB().Ping()
	if err != nil {
		panic(err)
	}

	return client
}
