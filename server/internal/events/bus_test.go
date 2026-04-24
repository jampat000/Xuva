package events

import (
	"context"
	"testing"
	"time"
)

func TestCancelIsIdempotentAndClosesSubscription(t *testing.T) {
	bus := NewBus(1)
	ch, cancel := bus.Subscribe(context.Background())

	cancel()
	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected subscription channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("subscription channel did not close")
	}
}

func TestPublishDoesNotBlockSlowSubscriber(t *testing.T) {
	bus := NewBus(1)
	ch, cancel := bus.Subscribe(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			bus.Publish("probe.progress", map[string]int{"index": i})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("publish blocked behind a slow subscriber")
	}

	select {
	case event := <-ch:
		if event.Type != "probe.progress" {
			t.Fatalf("expected probe.progress event, got %q", event.Type)
		}
	default:
		t.Fatal("expected at least one buffered event")
	}
}

func TestCloseClosesSubscribers(t *testing.T) {
	bus := NewBus(1)
	ch, cancel := bus.Subscribe(context.Background())
	defer cancel()

	bus.Close()
	bus.Publish("ignored", nil)

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected subscription channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("subscription channel did not close after bus close")
	}
}
