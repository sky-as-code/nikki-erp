package restful

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

// The container name in the REST struct's tag is a literal; it must be what the onion is
// registered under, or the handler resolves nothing at boot.
func TestUserEngineNameMatchesTheRegisteredOnion(t *testing.T) {
	assert.Equal(t, composable.EngineDependencyName(models.UserSchemaName), v1.UserEngineName)
}
