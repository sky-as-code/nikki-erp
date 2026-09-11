package composable

import (
	stdErr "errors"
	"sort"
	"sync"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The source index answers "which repository serves schema X" for the batched reads behind
// related and aggregate computed fields. Consumers never use it: an onion is reached through
// the dependency container by name. The index exists only because a computed field names its
// source schema as a string at read time, and dig cannot resolve a name it learns at runtime.

var sourceIndex = struct {
	mutex  sync.RWMutex
	repos  map[string]CrudRepository
	onions map[string]DynamicResourceEngineOnion
}{
	repos:  map[string]CrudRepository{},
	onions: map[string]DynamicResourceEngineOnion{},
}

func registerSource(schemaName string, repo CrudRepository, onion DynamicResourceEngineOnion) {
	sourceIndex.mutex.Lock()
	defer sourceIndex.mutex.Unlock()
	sourceIndex.repos[schemaName] = repo
	sourceIndex.onions[schemaName] = onion
}

// LookupSourceRepository returns the repository of a schema served by a composable onion. The
// legacy registry uses it as a fallback so a legacy resource may compute over a migrated one.
func LookupSourceRepository(schemaName string) (CrudRepository, bool) {
	sourceIndex.mutex.RLock()
	defer sourceIndex.mutex.RUnlock()
	repo, ok := sourceIndex.repos[schemaName]
	return repo, ok
}

// LookupSourceOnion returns the whole onion of a schema served by a composable engine, for a
// caller that needs the domain service of a sibling resource (import resolves and, on request,
// creates referenced records through it so their own rules run).
func LookupSourceOnion(schemaName string) (DynamicResourceEngineOnion, bool) {
	sourceIndex.mutex.RLock()
	defer sourceIndex.mutex.RUnlock()
	onion, ok := sourceIndex.onions[schemaName]
	return onion, ok
}

// BuiltSchemaNames lists every schema whose onion has been built so far, sorted. Onions are dig
// constructors, so one nobody resolved is absent; by OnAppStarted every routed onion is built.
func BuiltSchemaNames() []string {
	sourceIndex.mutex.RLock()
	defer sourceIndex.mutex.RUnlock()
	names := make([]string, 0, len(sourceIndex.onions))
	for name := range sourceIndex.onions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// AssertComputedFunctionsDefined checks every built onion's "function"-kind computed fields
// against the functions registered on it. Every onion is reported rather than only the first,
// so one boot surfaces the whole gap.
func AssertComputedFunctionsDefined() error {
	sourceIndex.mutex.RLock()
	onions := make([]DynamicResourceEngineOnion, 0, len(sourceIndex.onions))
	for _, onion := range sourceIndex.onions {
		onions = append(onions, onion)
	}
	sourceIndex.mutex.RUnlock()

	var failures []error
	for _, onion := range onions {
		if err := onion.AssertComputedFunctionsDefined(); err != nil {
			failures = append(failures, err)
		}
	}
	return joinErrors(failures)
}

var legacySourceResolver struct {
	mutex sync.RWMutex
	fn    SourceSearchFn
}

// SetLegacySourceResolver installs the fallback used when a computed field names a schema no
// composable onion serves, so a migrated resource may still compute over one that package
// engine serves. Package dynamicresource installs it at Init.
func SetLegacySourceResolver(fn SourceSearchFn) {
	legacySourceResolver.mutex.Lock()
	defer legacySourceResolver.mutex.Unlock()
	legacySourceResolver.fn = fn
}

// searchSourceRows is the batched read behind related computed fields: the rows of schemaName
// whose keyColumn is IN keys, projected down to fields. It goes through the source resource's
// own repository, so tenant/archive handling stays what a direct read would get.
func searchSourceRows(
	ctx corectx.Context, schemaName string, keyColumn string, keys []any, fields []string,
) ([]dmodel.DynamicFields, error) {
	repo, ok := LookupSourceRepository(schemaName)
	if !ok {
		legacySourceResolver.mutex.RLock()
		fallback := legacySourceResolver.fn
		legacySourceResolver.mutex.RUnlock()
		if fallback == nil {
			return nil, errors.Errorf("no resource engine for computed-field source '%s'", schemaName)
		}
		return fallback(ctx, schemaName, keyColumn, keys, fields)
	}
	return SearchRepositoryRows(ctx, repo, schemaName, keyColumn, keys, fields)
}

// SearchRepositoryRows runs the IN-keyed projection read a computed field needs against one
// repository. Exported so the legacy registry can serve a composable source the same way.
func SearchRepositoryRows(
	ctx corectx.Context, repo CrudRepository, schemaName string, keyColumn string, keys []any, fields []string,
) ([]dmodel.DynamicFields, error) {
	graph := dmodel.NewSearchGraph()
	graph.NewCondition(keyColumn, dmodel.In, keys...)

	found, err := repo.Search(ctx, dyn.RepoSearchParam{
		Fields: fields,
		Graph:  graph,
		Page:   0,
		Size:   len(keys),
	})
	if err != nil {
		return nil, err
	}
	if found != nil && found.ClientErrors.Count() > 0 {
		return nil, errors.Errorf(
			"computed-field source read of '%s' failed: %v", schemaName, found.ClientErrors.ToError())
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data.Items, nil
}

func joinErrors(errs []error) error {
	return stdErr.Join(errs...)
}
