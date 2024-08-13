package cancel

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	d := 100 * time.Millisecond

	t.Run("context is done before waiting", func(t *testing.T) {
		_, _, cancel, wait := New(ctx, d)
		defer cancel()

		err := wait()

		if err == nil {
			t.Errorf("resChan should not be nil")
		}
	})

	t.Run("context is done after waiting", func(t *testing.T) {
		_, _, cancel, wait := New(ctx, d)
		defer cancel()

		time.Sleep(200 * time.Millisecond)

		err := wait()

		if err == nil {
			t.Errorf("resChan should not be nil")
		}
	})

	t.Run("context is timeout", func(t *testing.T) {
		_, _, cancel, wait := New(ctx, d)
		defer cancel()

		time.Sleep(200 * time.Millisecond)

		err := wait()

		if err == nil {
			t.Errorf("error should not be nil")
		}
	})

	t.Run("chan must have a error", func(t *testing.T) {
		_, _, cancel, wait := New(ctx, d)
		defer cancel()

		err := wait()

		if err == nil {
			t.Errorf("error should not be nil")
		}
	})

	t.Run("must succeed", func(t *testing.T) {
		_, resChan, cancel, wait := New(ctx, d)
		defer cancel()

		go func() {
			resChan <- nil
		}()

		err := wait()

		if err != nil {
			t.Errorf("error should be nil")
		}
	})
}
func TestRunWithTimeout(t *testing.T) {
	ctx := context.Background()
	d := 100 * time.Millisecond

	t.Run("function completes within timeout", func(t *testing.T) {
		err := RunWithTimeout(ctx, d, func(ctx context.Context) error {
			// Simulate some work that completes within the timeout
			time.Sleep(50 * time.Millisecond)
			return nil
		})

		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("function exceeds timeout", func(t *testing.T) {
		err := RunWithTimeout(ctx, d, func(ctx context.Context) error {
			// Simulate some work that exceeds the timeout
			time.Sleep(200 * time.Millisecond)
			return nil
		})

		if err == nil {
			t.Errorf("expected an error, got nil")
		}
	})

	t.Run("function returns an error", func(t *testing.T) {
		expectedErr := errors.New("something went wrong")

		err := RunWithTimeout(ctx, d, func(ctx context.Context) error {
			// Simulate some work that returns an error
			return expectedErr
		})

		if err != expectedErr {
			t.Errorf("expected error: %v, got: %v", expectedErr, err)
		}
	})
}
