package testx

import (
	"context"
	"time"
)

func SetTimeout(ctx context.Context, timeout time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(timeout):
		panic("test timed out")
	}
}
