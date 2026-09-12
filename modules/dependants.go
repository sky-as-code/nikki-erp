package modules

import (
	"sort"

	"go.bryk.io/pkg/errors"
)

// ModuleDependantRegistry answers "who references this module's resources", which a module asks
// before deleting one of them. It is built once at startup from every module implementing
// InCodeModuleDependants and is read-only afterwards.
//
// Absent dependants are kept out of the answer rather than rejected. A module's dependant list is
// written once and ships in every binary, but the binaries load different module sets — coremart
// runs the vending machine, nikkierp does not — so naming a module that this binary did not load
// is a legitimate state, not a misconfiguration. Only the names that are actually loaded can
// answer a usage check, and only those are returned.
type ModuleDependantRegistry struct {
	dependants map[string][]string
	absent     map[string][]string
}

// NewModuleDependantRegistry collects the dependant lists and validates them against the loaded
// set. It rejects what is certainly wrong — a module naming itself, or naming the same dependant
// twice — and reports what is merely unusable here, so the caller can log it.
func NewModuleDependantRegistry(mods []InCodeModule) (*ModuleDependantRegistry, error) {
	loaded := make(map[string]bool, len(mods))
	for _, mod := range mods {
		loaded[mod.Name()] = true
	}

	registry := &ModuleDependantRegistry{
		dependants: make(map[string][]string),
		absent:     make(map[string][]string),
	}

	for _, mod := range mods {
		modWithDependants, ok := mod.(InCodeModuleDependants)
		if !ok {
			continue
		}
		present, absent, err := validateDependantList(mod.Name(), modWithDependants.Dependants(), loaded)
		if err != nil {
			return nil, err
		}
		registry.dependants[mod.Name()] = present
		if len(absent) > 0 {
			registry.absent[mod.Name()] = absent
		}
	}

	return registry, nil
}

func validateDependantList(
	owner string, declared []string, loaded map[string]bool,
) (present []string, absent []string, err error) {
	seen := make(map[string]bool, len(declared))
	present = make([]string, 0, len(declared))

	for _, dependant := range declared {
		if dependant == owner {
			return nil, nil, errors.Errorf(
				"module '%s' lists itself as its own dependant", owner)
		}
		if seen[dependant] {
			return nil, nil, errors.Errorf(
				"module '%s' lists dependant '%s' more than once", owner, dependant)
		}
		seen[dependant] = true

		if loaded[dependant] {
			present = append(present, dependant)
		} else {
			absent = append(absent, dependant)
		}
	}
	return present, absent, nil
}

// GetModuleDependants returns the loaded modules that may reference moduleName's resources.
// An unknown module and one with no dependants both answer empty, which is the same instruction
// to the caller: there is nobody to ask.
func (this *ModuleDependantRegistry) GetModuleDependants(moduleName string) []string {
	declared := this.dependants[moduleName]
	result := make([]string, len(declared))
	copy(result, declared)
	return result
}

// AbsentDependants returns the names moduleName declared that this binary did not load. It exists
// so startup can log them: the list is not an error, but a name that is absent because it was
// misspelled looks exactly like one that is absent by design, and the log is what makes the
// difference visible.
func (this *ModuleDependantRegistry) AbsentDependants(moduleName string) []string {
	declared := this.absent[moduleName]
	result := make([]string, len(declared))
	copy(result, declared)
	return result
}

// OwnersWithAbsentDependants lists, in a stable order, the modules that declared a dependant this
// binary did not load.
func (this *ModuleDependantRegistry) OwnersWithAbsentDependants() []string {
	owners := make([]string, 0, len(this.absent))
	for owner := range this.absent {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	return owners
}
