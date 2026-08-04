package interfaces

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
	NewEngine(schemaName string) (DynamicResourceEngine, error)

	// GetEngine returns the engine registered for the given schema name.
	GetEngine(schemaName string) (DynamicResourceEngine, bool)

	// MustGetEngine is GetEngine that panics when the engine is missing.
	MustGetEngine(schemaName string) DynamicResourceEngine

	// AllEngines returns every registered engine.
	AllEngines() []DynamicResourceEngine
}
