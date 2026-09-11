package services

import (
	"sort"
	"sync"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The resource hub is how one Sales service reaches another resource's repository or domain
// service at call time. The resources of this module read each other in cycles (an order reads its
// bills, a bill reads the order it settles), so the layers cannot be handed to each other at
// construction; the dynamicengines package installs every built onion here instead, and a service
// looks a peer up when it needs it.
//
// It is module-private: nothing outside Sales reaches a resource this way. Other modules inject
// the typed ports Sales publishes into the dependency container.

type resourceEntry struct {
	repo   composable.CrudRepository
	domSvc composable.CrudDomainService
}

var resourceHub = struct {
	mutex   sync.RWMutex
	entries map[string]resourceEntry
}{entries: map[string]resourceEntry{}}

// InstallResource records a built onion's repository and domain service under its schema name.
// Called by the module's engine registration once the container has built the onion.
func InstallResource(schemaName string, repo composable.CrudRepository, domSvc composable.CrudDomainService) {
	resourceHub.mutex.Lock()
	defer resourceHub.mutex.Unlock()
	resourceHub.entries[schemaName] = resourceEntry{repo: repo, domSvc: domSvc}
}

// InstalledResourceNames lists the schemas installed so far, sorted, for the boot-time check that
// every resource the services may reach was built.
func InstalledResourceNames() []string {
	resourceHub.mutex.RLock()
	defer resourceHub.mutex.RUnlock()
	names := make([]string, 0, len(resourceHub.entries))
	for name := range resourceHub.entries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func lookupResource(schemaName string) (resourceEntry, error) {
	resourceHub.mutex.RLock()
	defer resourceHub.mutex.RUnlock()
	entry, ok := resourceHub.entries[schemaName]
	if !ok {
		return resourceEntry{}, errors.Errorf("no sales resource installed for '%s'", schemaName)
	}
	return entry, nil
}

// repoFor resolves another resource's repository. It is a variable so a test can substitute its
// own repositories: the hub is filled during Init, which a unit test cannot run.
var repoFor = func(schemaName string) (composable.CrudRepository, error) {
	entry, err := lookupResource(schemaName)
	if err != nil {
		return nil, err
	}
	return entry.repo, nil
}

// domainServiceFor resolves another resource's domain service, for the few writes that must go
// through a resource's create pipeline (schema defaults, audit fields) rather than its repository.
//
// It has no legacy fallback: no Sales service reaches a peer's domain service today, so a lookup
// that misses is a wiring error rather than a not-yet-migrated resource.
var domainServiceFor = func(schemaName string) (composable.CrudDomainService, error) {
	entry, err := lookupResource(schemaName)
	if err != nil {
		return nil, err
	}
	return entry.domSvc, nil
}

// RepositoryFor exposes the hub lookup to sibling packages that need another resource's
// repository. It answers whichever generation serves the schema, so a caller works unchanged
// across the migration -- which EngineFor cannot do, because a migrated schema has no legacy
// engine and its lookup fails with "no resource engine for '<schema>'".
func RepositoryFor(schemaName string) (composable.CrudRepository, error) {
	return repoFor(schemaName)
}

// DomainServiceFor exposes the hub's domain-service lookup to sibling packages, for the few
// writes that must go through a resource's create pipeline -- schema defaults and audit fields --
// rather than its repository.
func DomainServiceFor(schemaName string) (composable.CrudDomainService, error) {
	return domainServiceFor(schemaName)
}

// The typed peer lookups. Each answers the derived service a sibling package needs without that
// package importing dynamicengines, which would close an import cycle: dynamicengines already
// imports app to build the onions.

// FulfillmentMethodService answers the derived method service, or nil when the resource is not
// built. Nil rather than an error: resolution then treats the sale as an ordinary one, so an
// order that named no method is not refused because a machine-dispensing feature is unconfigured.
func FulfillmentMethodService() *SalesFulfillmentMethodDomainServiceImpl {
	entry, err := lookupResource(models.SalesFulfillmentMethodSchemaName)
	if err != nil {
		return nil
	}
	derived, _ := entry.domSvc.(*SalesFulfillmentMethodDomainServiceImpl)
	return derived
}

// SalesPointService answers the derived point service. An error rather than nil: every caller
// needs it to do its work, so a missing one is a wiring fault worth reporting as such.
func SalesPointService() (*SalesPointDomainServiceImpl, error) {
	entry, err := lookupResource(models.SalesPointSchemaName)
	if err != nil {
		return nil, err
	}
	derived, ok := entry.domSvc.(*SalesPointDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the sales point onion is not built with the derived point service; " +
				"the module's engine registration must install NewSalesPointDomainService")
	}
	return derived, nil
}
