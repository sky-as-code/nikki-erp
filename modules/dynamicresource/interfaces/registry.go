package interfaces

// NewEngineOptions customizes an engine at creation time.
// The zero value is valid and reproduces the default engine behavior.
type NewEngineOptions struct {
	// DefaultSearchFields is the field list a search returns when it specifies neither
	// fields nor a resolvable view. When empty, every column of the schema is returned.
	//
	// Primary key fields are always included by the query builder, so listing them here
	// is redundant.
	DefaultSearchFields []string
}

// DynamicResourceEngineRegistry owns every resource engine of the running application.
// It is a process-wide singleton, reached through dynamicresource.Registry().
//
// Future work: part of the registry data (engine list and their action definitions) is
// meant to be loadable from the database, so that a resource can be declared without a
// code change. The seam is EngineFactory: a loader would call NewEngine followed by
// DefineAction for each persisted definition. That loading is deliberately not implemented yet.
type DynamicResourceEngineRegistry interface {
	// NewEngine creates and registers an engine for the given dynamic-model schema name.
	// It fails when the schema is unknown or an engine for it already exists.
	// Only the first NewEngineOptions is used, the rest are ignored.
	NewEngine(schemaName string, options ...NewEngineOptions) (DynamicResourceEngine, error)

	// GetEngine returns the engine registered for the given schema name.
	GetEngine(schemaName string) (DynamicResourceEngine, bool)

	// MustGetEngine is GetEngine that panics when the engine is missing.
	MustGetEngine(schemaName string) DynamicResourceEngine

	// AllEngines returns every registered engine.
	AllEngines() []DynamicResourceEngine
}
