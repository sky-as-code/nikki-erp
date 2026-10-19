package composable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

type stubEngine struct {
	DynamicResourceEngineOnion
	name string
}

func TestResolveEngines_BySchemaName(t *testing.T) {
	for _, name := range []string{"lookup_test_a", "lookup_test_b"} {
		engine := &stubEngine{name: name}
		require.NoError(t, deps.RegisterNamed(EngineDependencyName(name), func() DynamicResourceEngineOnion {
			return engine
		}))
	}

	engines, err := ResolveEngines("lookup_test_b", "lookup_test_a")

	require.NoError(t, err)
	require.Len(t, engines, 2)
	assert.Equal(t, "lookup_test_b", engines[0].(*stubEngine).name)
	assert.Equal(t, "lookup_test_a", engines[1].(*stubEngine).name)
}

func TestResolveEngines_UnknownSchema(t *testing.T) {
	_, err := ResolveEngines("lookup_test_missing")
	assert.Error(t, err)
}
