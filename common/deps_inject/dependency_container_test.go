package deps_inject

import (
	"fmt"
	"sync"
	"testing"
	"time"

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

type concurrentService struct {
	Calls int
}

type concurrentLeaf struct{}

// dig is not safe for concurrent use: without the lock this trips "concurrent map read and map
// write" inside the container. Each goroutine also invokes re-entrantly, the way a constructor
// that resolves its own dependencies does.
func TestInvokeFromManyGoroutines(t *testing.T) {
	assert.NoError(t, Register(func() *concurrentLeaf { return &concurrentLeaf{} }))
	assert.NoError(t, Register(func(_ *concurrentLeaf) *concurrentService {
		nested := Invoke(func(_ *concurrentLeaf) {})
		if nested != nil {
			panic(nested)
		}
		return &concurrentService{}
	}))

	const workers = 50
	errs := make(chan error, 2*workers)
	var wait sync.WaitGroup
	for i := 0; i < workers; i++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			errs <- Invoke(func(svc *concurrentService) { svc.Calls++ })
			errs <- RegisterNamed(fmt.Sprintf("concurrent_%d", index), func() *plainService {
				return &plainService{Name: "concurrent"}
			})
		}(i)
	}
	wait.Wait()
	close(errs)

	for err := range errs {
		assert.NoError(t, err)
	}
	var resolved *concurrentService
	assert.NoError(t, Invoke(func(svc *concurrentService) { resolved = svc }))
	assert.Equal(t, workers, resolved.Calls, "the singleton is built once and every Invoke is serialized")
}

func TestReentrantMutexReleasesOnlyAtTheOutermostUnlock(t *testing.T) {
	var mutex reentrantMutex
	mutex.Lock()
	mutex.Lock()
	mutex.Unlock()

	acquired := make(chan struct{})
	go func() {
		mutex.Lock()
		close(acquired)
		mutex.Unlock()
	}()
	select {
	case <-acquired:
		t.Fatal("another goroutine acquired the mutex while it was still held")
	case <-time.After(50 * time.Millisecond):
	}
	mutex.Unlock()
	<-acquired
}
