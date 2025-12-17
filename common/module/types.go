package module

import (
	dschema "github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"github.com/sky-as-code/nikki-erp/common/semver"
)

type ModuleInitOptions struct {
	RegisterSchema dschema.RegisterSchemaFunc
}

type InCodeModule interface {
	Deps() []string
	// LabelKey is the translation key.
	LabelKey() string
	Name() string
	Init(opts ModuleInitOptions) error
	Version() semver.SemVer
}

type InCodeModuleAppStarted interface {
	OnAppStarted() error
}
