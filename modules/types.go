package modules

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

type NikkiModule interface {
	Name() string
	Init() error
	Deps() []string
}
