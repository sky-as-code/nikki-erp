package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
)

type depsTestModule struct {
	name string
	deps []string
}

func (this *depsTestModule) Deps() []string          { return this.deps }
func (this *depsTestModule) LabelKey() string        { return this.name + ".moduleLabel" }
func (this *depsTestModule) Name() string            { return this.name }
func (this *depsTestModule) Init() error             { return nil }
func (this *depsTestModule) IsInternal() bool        { return false }
func (this *depsTestModule) ModelPrefixes() []string { return []string{this.name} }
func (this *depsTestModule) HasUi() bool             { return false }
func (this *depsTestModule) Category() modules.ModuleCategory {
	return modules.ModuleCategoryFoundation
}
func (this *depsTestModule) Version() semver.SemVer { return *semver.MustParseSemVer("v1.0.0") }

func TestComputeDepLevels(t *testing.T) {
	levels, err := computeDepLevels([]modules.InCodeModule{
		&depsTestModule{name: "sales", deps: []string{"inventory", "essential"}},
		&depsTestModule{name: "inventory", deps: []string{"essential", "not_loaded"}},
		&depsTestModule{name: "essential", deps: []string{"core"}},
		&depsTestModule{name: "core"},
	})

	require.NoError(t, err)
	assert.Equal(t, map[string]int32{"core": 0, "essential": 1, "inventory": 2, "sales": 3}, levels)
}

func TestComputeDepLevels_Cycle(t *testing.T) {
	_, err := computeDepLevels([]modules.InCodeModule{
		&depsTestModule{name: "a", deps: []string{"b"}},
		&depsTestModule{name: "b", deps: []string{"a"}},
	})

	assert.ErrorContains(t, err, "cycle")
}
