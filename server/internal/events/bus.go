package events

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Event struct {
	ID        uint64    `json:"id"`
	Type      string    `json:"type"`
	Data      any       `json:"data,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type Bus struct {
	buffer int
	nextID atomic.Uint64

	mu          sync.RWMutex
	closed      bool
	subscribers map[chan Event]struct{}
}

func NewBus(buffer int) *Bus {
	if buffer < 1 {
		buffer = 1
	}
	return &Bus{
		buffer:      buffer,
		subscribers: make(map[chan Event]struct{}),
	}
}

func (b *Bus) Subscribe(ctx context.Context) (<-chan Event, func()) {
	ch := make(chan Event, b.buffer)

	b.mu.Lock()
	if b.closed {
		close(ch)
		b.mu.Unlock()
		return ch, func() {}
	}
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	stop := make(chan struct{})
	var once sync.Once
	cancel := func() {
		once.Do(func() {
			b.mu.Lock()
			if _, ok := b.subscribers[ch]; ok {
				delete(b.subscribers, ch)
				close(ch)
			}
			b.mu.Unlock()
			close(stop)
		})
	}

	go func() {
		select {
		case <-ctx.Done():
			cancel()
		case <-stop:
		}
	}()

	return ch, cancel
}

func (b *Bus) Publish(eventType string, data any) {
	event := Event{
		ID:        b.nextID.Add(1),
		Type:      eventType,
		Data:      data,
		CreatedAt: time.Now().UTC(),
	}

	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return
	}
	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	for ch := range b.subscribers {
		close(ch)
		delete(b.subscribers, ch)
	}
}
