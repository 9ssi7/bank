package cancel

import (
	"context"
	"time"
)

// WaitFunc is a function that waits for the context to be done.
// It is used to wait for the context to be done before returning.
type WaitFunc func() error

// TimeoutedFunc is a function that takes a context and returns an error.
// It is used to run a function with a timeout.
type TimeoutedFunc func(ctx context.Context) error

// New creates a new context with a timeout and returns a channel to send the result and a function to cancel the context.
// It also returns a function that waits for the context to be done before returning.
func New(ctx context.Context, d time.Duration) (context.Context, chan error, context.CancelFunc, WaitFunc) {
	resChan := make(chan error)
	ctx, cancel := context.WithTimeout(ctx, d)
	return ctx, resChan, cancel, func() error {
		return Wait(ctx, resChan)
	}
}

// Wait waits for the context to be done.
func Wait(ctx context.Context, ch chan error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-ch:
		return err
	}
}

// NewWithTimeout creates a new context with a timeout and runs the given function with the context.
// It returns an error if the function returns an error or if the context times out.
func NewWithTimeout(ctx context.Context, d time.Duration, f TimeoutedFunc) error {
	ctx, resChan, cancel, waiter := New(ctx, d)
	defer cancel()
	go func() {
		resChan <- f(ctx)
	}()
	return waiter()
}
