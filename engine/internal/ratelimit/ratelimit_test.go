package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestUnlimitedReturnsImmediately(t *testing.T) {
	l := New(0)
	start := time.Now()
	if err := l.WaitN(context.Background(), 1<<20); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Fatalf("unlimited limiter should not block")
	}
}

func TestLimitedAllowsWithinBudget(t *testing.T) {
	l := New(1024) // 1 KB/s, full bucket on creation
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := l.WaitN(ctx, 256); err != nil {
		t.Fatalf("expected to consume within budget: %v", err)
	}
}

func TestContextCancelInterrupts(t *testing.T) {
	l := New(100)
	l.mu.Lock()
	l.tokens = 0
	l.last = time.Now()
	l.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- l.WaitN(ctx, 100)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("rate-limit wait did not stop after context cancellation")
	}
}

func TestCanceledContextDoesNotConsumeAvailableTokens(t *testing.T) {
	l := New(1024)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := l.WaitN(ctx, 1); err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
