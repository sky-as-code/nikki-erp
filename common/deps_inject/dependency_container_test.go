package deps_inject

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
)

// The container is a package-level singleton shared by all tests in this package,
// so each test uses its own types to avoid cross-test collisions.

type namedEngine struct {
	Id string
}

type namedEngineParam struct {
	dig.In

	UserEngine *namedEngine `name:"engine_iam_user"`
	OrgEngine  *namedEngine `name:"engine_iam_org"`
}

func TestRegisterNamed(t *testing.T) {
	err := RegisterNamed("engine_iam_user", func() *namedEngine {
		return &namedEngine{Id: "iam_user"}
	})
	assert.NoError(t, err)

	err = RegisterNamed("engine_iam_org", func() *namedEngine {
		return &namedEngine{Id: "iam_org"}
	})
	assert.NoError(t, err)

	var userEngine, orgEngine *namedEngine
	err = Invoke(func(param namedEngineParam) {
		userEngine = param.UserEngine
		orgEngine = param.OrgEngine
	})

	assert.NoError(t, err)
	assert.Equal(t, "iam_user", userEngine.Id)
	assert.Equal(t, "iam_org", orgEngine.Id)
}

type unnamedOnly struct{}

func TestRegisterNamedIsNotResolvableWithoutName(t *testing.T) {
	err := RegisterNamed("engine_unnamed_only", func() *unnamedOnly {
		return &unnamedOnly{}
	})
	assert.NoError(t, err)

	err = Invoke(func(_ *unnamedOnly) {})
	assert.Error(t, err, "a named value must not satisfy an unnamed dependency")
}

type plainService struct {
	Name string
}

func TestRegisterPlainIsUnaffected(t *testing.T) {
	err := Register(func() *plainService {
		return &plainService{Name: "plain"}
	})
	assert.NoError(t, err)

	var resolved *plainService
	err = Invoke(func(svc *plainService) {
		resolved = svc
	})

	assert.NoError(t, err)
	assert.Equal(t, "plain", resolved.Name)
}
