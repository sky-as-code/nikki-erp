package services

import (
	"sort"
	"sync"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The resource hub is how the payment flow reaches a resource's repository or domain service at
// call time. The order service is itself an argument of the order onion's application service, so
// handing it the onion's repository at construction would be a cycle; the dynamicengines package
// installs every built onion here instead, and a service looks a peer up when it needs it.
//
// It is module-private: other modules inject the ports in interfaces/ instead.

type resourceEntry struct {
	repo   composable.CrudRepository
	domSvc composable.CrudDomainService
}

var resourceHub = struct {
	mutex   sync.RWMutex
	entries map[string]resourceEntry
}{entries: map[string]resourceEntry{}}

// InstallResource records a built onion's repository and domain service under its schema name.
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
		return resourceEntry{}, errors.Errorf("no paymentinvoice resource installed for '%s'", schemaName)
	}
	return entry, nil
}

// repoFor resolves a resource's repository. It is a variable so a test can substitute its own:
// the hub is filled during Init, which a unit test cannot run.
var repoFor = func(schemaName string) (composable.CrudRepository, error) {
	entry, err := lookupResource(schemaName)
	if err != nil {
		return nil, err
	}
	return entry.repo, nil
}

// domainServiceFor resolves a resource's domain service, for the writes that must go through the
// create pipeline (schema defaults, audit fields) rather than the bare repository.
var domainServiceFor = func(schemaName string) (composable.CrudDomainService, error) {
	entry, err := lookupResource(schemaName)
	if err != nil {
		return nil, err
	}
	return entry.domSvc, nil
}
