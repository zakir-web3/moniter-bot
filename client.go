package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

const maxRetries = 3

func retry[T any](ctx context.Context, name string, fn func(ctx context.Context) (T, error)) (T, error) {
	var zero T
	var lastErr error
	for i := range maxRetries {
		if i > 0 {
			backoff := time.Duration(1<<uint(i-1)) * time.Second
			slog.Warn("retrying", "operation", name, "attempt", i+1, "backoff", backoff, "lastError", lastErr)
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(backoff):
			}
		}
		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return zero, lastErr
}

func retryDo(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	_, err := retry(ctx, name, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	})
	return err
}
