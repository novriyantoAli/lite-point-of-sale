package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"
)

func TestRunCallsTheJobThenStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	calls := 0
	run := func(context.Context) error {
		mu.Lock()
		calls++
		mu.Unlock()
		return nil
	}

	done := make(chan struct{})
	go func() {
		Run(ctx, run, 5*time.Millisecond, slog.New(slog.DiscardHandler))
		close(done)
	}()

	// The immediate call, then one after the first tick.
	waitFor(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return calls >= 2
	})

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not stop after the context was cancelled")
	}
}

func TestRunKeepsGoingAfterAnError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	calls := 0
	run := func(context.Context) error {
		mu.Lock()
		calls++
		mu.Unlock()
		return context.Canceled // any error: the loop must not stop
	}

	done := make(chan struct{})
	go func() {
		Run(ctx, run, 5*time.Millisecond, slog.New(slog.DiscardHandler))
		close(done)
	}()

	// Two calls despite the first erroring: the job ran, failed, and ran again.
	waitFor(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return calls >= 2
	})

	cancel()
	<-done
}

func waitFor(t *testing.T, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}

	t.Fatal("condition did not become true in time")
}
