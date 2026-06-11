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
	l := New(100) // max bucket = 100 bytes
	_ = l.WaitN(context.Background(), 80) // drain most of the bucket
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := l.WaitN(ctx, 80); err == nil {
		t.Fatalf("expected context cancellation to interrupt the wait")
	}
}
