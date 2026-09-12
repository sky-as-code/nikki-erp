package external

import (
	"sync"

	"github.com/sky-as-code/nikki-erp/common/model"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// streamKey identifies one person's streams on this instance. The organization is part of the key
// because the same person in two organizations has two separate inboxes, and an event must never
// cross between them (BR-FS 4).
type streamKey struct {
	orgId  model.Id
	userId model.Id
}

// MemoryStreamRegistry tracks the streams this instance is serving.
//
// In-memory by design: a stream is a live HTTP response and cannot be handed to another process.
// The broker is what makes the instances behave as one, so this being per-instance is not the
// limitation it looks like.
type MemoryStreamRegistry struct {
	mutex   sync.RWMutex
	streams map[streamKey]map[int]chan it.StreamEvent
	nextId  int
}

func NewMemoryStreamRegistry() it.StreamRegistry {
	return &MemoryStreamRegistry{
		streams: make(map[streamKey]map[int]chan it.StreamEvent),
	}
}

// Register adds a stream and returns it with the function that removes it.
//
// One person may hold several at once — two browser tabs are two streams — so the streams of a key
// are a set rather than a single channel. Returning the remover instead of exposing an Unregister
// keyed by id keeps the caller from having to track the id itself, and makes `defer remove()` the
// obvious thing to write.
func (this *MemoryStreamRegistry) Register(
	orgId model.Id, userId model.Id, buffer int,
) (<-chan it.StreamEvent, func()) {
	if buffer <= 0 {
		buffer = 1
	}

	key := streamKey{orgId: orgId, userId: userId}
	events := make(chan it.StreamEvent, buffer)

	this.mutex.Lock()
	this.nextId++
	id := this.nextId
	if this.streams[key] == nil {
		this.streams[key] = make(map[int]chan it.StreamEvent)
	}
	this.streams[key][id] = events
	this.mutex.Unlock()

	return events, func() { this.remove(key, id) }
}

func (this *MemoryStreamRegistry) remove(key streamKey, id int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	streams, found := this.streams[key]
	if !found {
		return
	}
	events, found := streams[id]
	if !found {
		return
	}

	delete(streams, id)
	if len(streams) == 0 {
		delete(this.streams, key)
	}
	// Closing tells the handler to finish. Safe because only this registry sends on the channel,
	// and remove is idempotent through the found checks above.
	close(events)
}

// Dispatch offers an event to each of the user's streams on this instance.
//
// It never blocks. A stream whose buffer is full is a client that is not reading fast enough, and
// the contract is to drop rather than to wait (BR-FS 19): blocking here would let one suspended
// browser tab stall the dispatcher for every other user on the instance. The dropped client is not
// harmed — the notification is committed, and it replays from after_seq on its next connection.
func (this *MemoryStreamRegistry) Dispatch(
	orgId model.Id, userId model.Id, event it.StreamEvent,
) int {
	key := streamKey{orgId: orgId, userId: userId}

	this.mutex.RLock()
	defer this.mutex.RUnlock()

	delivered := 0
	for _, events := range this.streams[key] {
		select {
		case events <- event:
			delivered++
		default:
		}
	}
	return delivered
}

func (this *MemoryStreamRegistry) HasStreams(orgId model.Id, userId model.Id) bool {
	key := streamKey{orgId: orgId, userId: userId}

	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return len(this.streams[key]) > 0
}

func (this *MemoryStreamRegistry) Count() int {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	total := 0
	for _, streams := range this.streams {
		total += len(streams)
	}
	return total
}
