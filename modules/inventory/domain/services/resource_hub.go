package services

import (
	"sort"
	"sync"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The resource hub is how one Inventory service reaches another resource's repository or domain
// service at call time. The resources of this module read each other in cycles (a location
// consults the quants it holds, a quant summary reads its locations), so the layers cannot be
// handed to each other at construction; the dynamicengines package installs every built onion
// here instead, and a service looks a peer up when it needs it.
//
// It is module-private: nothing outside Inventory reaches a resource this way. Other modules
// inject the typed ports Inventory publishes into the dependency container.

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
		return resourceEntry{}, errors.Errorf("no inventory resource installed for '%s'", schemaName)
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
var domainServiceFor = func(schemaName string) (composable.CrudDomainService, error) {
	entry, err := lookupResource(schemaName)
	if err != nil {
		return nil, err
	}
	return entry.domSvc, nil
}
