package modules

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/semver"
)

type JSON map[string]any

type Auditable interface {
	SetCreatedUtcNow()
	SetUpdatedUtcNow()
}

type Copiable interface {
	CopyTo(destPtr any)
	CopyFrom(sourcePtr any) any
}

type SoftDeletable interface {
	SetDeletedUtcNow()
}

type ValueObject interface {
	// Value() (any, error)
	Json() any
	String() string
}

type DomainModel interface {
	// Copiable

	PrimaryKey() ValueObject
	Clone() any
	Validate(forEdit bool) error
}

type InCodeModule interface {
	Deps() []string
	// Deprecated: Use Name() instead.
	LabelKey() string
	Name() string
	Init() error
	IsInternal() bool
	Version() semver.SemVer
	// ModelPrefix returns the prefix this module's schemas are named with, without a
	// trailing separator - GenSql appends the "_" itself.
	ModelPrefix() string
}

type DynamicModule interface {
	Deps() []string
	// LabelKey is the translation key.
	LabelKey() string
	Name() string
	Init() error
	IsInternal() bool
	RegisterModels() error
	Version() semver.SemVer
	// ModelPrefix returns the prefix this module's schemas are named with, without a
	// trailing separator - GenSql appends the "_" itself.
	ModelPrefix() string
}

type InCodeModuleAppStarted interface {
	OnAppStarted() error
}

// InCodeModuleDependants is the optional counterpart to Deps(): the modules that reference THIS
// module's resources, declared by this module rather than by them.
//
// The two lists answer opposite questions and neither can be derived from the other. Deps() says
// what this module needs in order to start; Dependants() says who must be consulted before one of
// its resources is deleted. A dependency graph cannot answer the second question — inventory
// depending on essential says nothing about who depends on inventory — so the list is hard-coded
// and owned by the module whose resources are at stake.
//
// The names are canonical module names, matching Name(). A name that is not loaded in this binary
// is skipped rather than rejected: the same module ships in binaries with different module sets,
// and a resource whose consumer is absent has no consumer to ask.
type InCodeModuleDependants interface {
	Dependants() []string
}

// InCodeModuleAppStopping is an optional hook, the counterpart to InCodeModuleAppStarted,
// invoked after the OS termination signal and before the HTTP server shuts down. A module
// that owns background goroutines implements it to stop accepting new work and drain what
// is already in flight.
//
// It is best-effort. Every module shares one deadline carried by ctx, and a module that
// blocks past it is abandoned rather than allowed to hold up the shutdown of the rest. A
// module whose work must survive an abrupt stop has to be recoverable from persisted state;
// this hook shortens the window in which that recovery is needed, it does not remove it.
type InCodeModuleAppStopping interface {
	OnAppStopping(ctx context.Context) error
}

type DefaultConfig []byte
