package modules

import "github.com/sky-as-code/nikki-erp/common/semver"

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
	// LabelKey is the translation key.
	LabelKey() string
	Name() string
	Init() error
	Version() semver.SemVer
}

type InCodeModuleAppStarted interface {
	OnAppStarted() error
}
