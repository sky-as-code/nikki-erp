package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/semver"
)

// plainModule is a module that declares no dependants, which most modules do. It is the base
// shape; ownerModule adds Dependants() on top, so the two differ by exactly the optional
// interface the registry type-asserts.
type plainModule struct {
	name string
}

func (this *plainModule) Name() string           { return this.name }
func (this *plainModule) Deps() []string         { return nil }
func (this *plainModule) LabelKey() string       { return this.name }
func (this *plainModule) Init() error            { return nil }
func (this *plainModule) IsInternal() bool       { return false }
func (this *plainModule) ModelPrefix() string    { return this.name }
func (this *plainModule) Version() semver.SemVer { return *semver.MustParseSemVer("v1.0.0") }

type ownerModule struct {
	plainModule
	dependants []string
}

func (this *ownerModule) Dependants() []string { return this.dependants }

func owner(name string, dependants ...string) InCodeModule {
	return &ownerModule{plainModule: plainModule{name: name}, dependants: dependants}
}

func plain(name string) InCodeModule {
	return &plainModule{name: name}
}

func TestDependantRegistry_ReturnsLoadedDependants(t *testing.T) {
	registry, err := NewModuleDependantRegistry([]InCodeModule{
		owner("inventory", "sales", "purchase"),
		plain("sales"),
		plain("purchase"),
	})
	require.NoError(t, err)

	assert.Equal(t, []string{"sales", "purchase"}, registry.GetModuleDependants("inventory"))
	assert.Empty(t, registry.AbsentDependants("inventory"))
}

// A module with no dependants and an unknown module answer the same way, because the caller does
// the same thing with both: ask nobody.
func TestDependantRegistry_EmptyForModuleWithoutDependants(t *testing.T) {
	registry, err := NewModuleDependantRegistry([]InCodeModule{plain("sales")})
	require.NoError(t, err)

	assert.Empty(t, registry.GetModuleDependants("sales"))
	assert.Empty(t, registry.GetModuleDependants("nope"))
}

// The case behind docs/problems/inventory/001: inventory names the vending machine, which only
// the coremart binary loads. Here it must not fail startup, and must not be dispatched to.
func TestDependantRegistry_SkipsDependantNotLoadedInThisBinary(t *testing.T) {
	registry, err := NewModuleDependantRegistry([]InCodeModule{
		owner("inventory", "sales", "vendingmachine"),
		plain("sales"),
	})
	require.NoError(t, err)

	assert.Equal(t, []string{"sales"}, registry.GetModuleDependants("inventory"),
		"an absent dependant cannot answer a usage check, so it is not dispatched to")
	assert.Equal(t, []string{"vendingmachine"}, registry.AbsentDependants("inventory"),
		"it is still reported, so startup can log what it skipped")
	assert.Equal(t, []string{"inventory"}, registry.OwnersWithAbsentDependants())
}

func TestDependantRegistry_RejectsSelfReference(t *testing.T) {
	_, err := NewModuleDependantRegistry([]InCodeModule{owner("inventory", "inventory")})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "lists itself")
}

func TestDependantRegistry_RejectsDuplicate(t *testing.T) {
	_, err := NewModuleDependantRegistry([]InCodeModule{
		owner("inventory", "sales", "sales"),
		plain("sales"),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "more than once")
}

// The registry hands out copies: a caller that sorts or filters the answer must not be able to
// reorder what the next caller sees.
func TestDependantRegistry_AnswerIsACopy(t *testing.T) {
	registry, err := NewModuleDependantRegistry([]InCodeModule{
		owner("inventory", "sales", "purchase"),
		plain("sales"),
		plain("purchase"),
	})
	require.NoError(t, err)

	answer := registry.GetModuleDependants("inventory")
	answer[0] = "tampered"

	assert.Equal(t, []string{"sales", "purchase"}, registry.GetModuleDependants("inventory"))
}
