package deps_inject

import (
	"bytes"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
)

// reentrantMutex lets the goroutine holding it lock it again, which a constructor running inside
// Invoke does whenever it registers or resolves in turn. Any other goroutine waits.
type reentrantMutex struct {
	mu    sync.Mutex
	owner atomic.Int64
	// depth is touched only by the owning goroutine.
	depth int
}

func (this *reentrantMutex) Lock() {
	goroutineId := currentGoroutineId()
	if this.owner.Load() == goroutineId {
		this.depth++
		return
	}
	this.mu.Lock()
	this.owner.Store(goroutineId)
	this.depth = 1
}

func (this *reentrantMutex) Unlock() {
	this.depth--
	if this.depth > 0 {
		return
	}
	this.owner.Store(0)
	this.mu.Unlock()
}

var goroutinePrefix = []byte("goroutine ")

// currentGoroutineId parses the id from the first line of the stack trace, "goroutine 123 [...".
// Go exposes no goroutine id on purpose; the stack header is the stable, documented-by-practice
// way to read it, and it costs about a microsecond, negligible against a dig resolution.
func currentGoroutineId() int64 {
	buffer := make([]byte, 64)
	buffer = buffer[:runtime.Stack(buffer, false)]
	buffer = bytes.TrimPrefix(buffer, goroutinePrefix)
	if end := bytes.IndexByte(buffer, ' '); end > 0 {
		buffer = buffer[:end]
	}
	goroutineId, err := strconv.ParseInt(string(buffer), 10, 64)
	if err != nil {
		panic("deps_inject: cannot read the goroutine id from the stack header")
	}
	return goroutineId
}
