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
}

type InCodeModuleAppStarted interface {
	OnAppStarted() error
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
