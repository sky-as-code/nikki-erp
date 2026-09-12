package external

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/common/model"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

const (
	orgA  = model.Id("org-a")
	orgB  = model.Id("org-b")
	userA = model.Id("user-a")
	userB = model.Id("user-b")
)

func event(seq int64) it.StreamEvent {
	return it.StreamEvent{Seq: &seq, Type: it.StreamEventNotificationCreated}
}

func TestDispatchReachesEveryStreamOfOneUser(t *testing.T) {
	registry := NewMemoryStreamRegistry()

	firstTab, closeFirst := registry.Register(orgA, userA, 4)
	defer closeFirst()
	secondTab, closeSecond := registry.Register(orgA, userA, 4)
	defer closeSecond()

	assert.Equal(t, 2, registry.Dispatch(orgA, userA, event(1)))
	assert.Equal(t, int64(1), *(<-firstTab).Seq)
	assert.Equal(t, int64(1), *(<-secondTab).Seq)
}

// A stream belongs to one person in one organization. The same user id in another organization is
// a different inbox and must never see the event (BR-FS 4, BR-FS 6).
func TestDispatchIsScopedToOrgAndUser(t *testing.T) {
	registry := NewMemoryStreamRegistry()

	_, closeOther := registry.Register(orgB, userA, 4)
	defer closeOther()
	_, closeOtherUser := registry.Register(orgA, userB, 4)
	defer closeOtherUser()

	assert.Equal(t, 0, registry.Dispatch(orgA, userA, event(1)))
}

// A client that stops reading must not stall the dispatcher. Once its buffer is full the extra
// events are dropped, and the client recovers them by reconnecting with after_seq (BR-FS 19).
func TestDispatchDropsRatherThanBlocksOnAFullBuffer(t *testing.T) {
	registry := NewMemoryStreamRegistry()

	events, remove := registry.Register(orgA, userA, 1)
	defer remove()

	assert.Equal(t, 1, registry.Dispatch(orgA, userA, event(1)))
	assert.Equal(t, 0, registry.Dispatch(orgA, userA, event(2)), "second event must be dropped")

	assert.Equal(t, int64(1), *(<-events).Seq)
}

func TestRemoveClosesTheStreamAndIsIdempotent(t *testing.T) {
	registry := NewMemoryStreamRegistry()

	events, remove := registry.Register(orgA, userA, 1)
	assert.Equal(t, 1, registry.Count())

	remove()
	remove()

	_, open := <-events
	assert.False(t, open, "removing a stream must close it")
	assert.Equal(t, 0, registry.Count())
	assert.Equal(t, 0, registry.Dispatch(orgA, userA, event(1)))
}
