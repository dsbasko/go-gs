package gogs

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	ShortDelay = 50 * time.Millisecond
	LongDelay  = time.Second
)

func Test_GracefulShutdown_Simple(t *testing.T) {
	t.Parallel()
	gs, _, _ := NewContext(context.Background(), syscall.SIGINT)

	gs.Subscribe()
	assert.Equal(t, int32(1), gs.Count())

	go func() {
		shortDelay()
		gs.Unsubscribe()
	}()

	gs.Wait()
	assert.Equal(t, int32(0), gs.Count())
}

func Test_GracefulShutdown_Context(t *testing.T) {
	t.Parallel()

	t.Run("Signal", func(t *testing.T) {
		var graceful bool
		gs, ctx, _ := NewContext(context.Background(), syscall.SIGINT)

		gs.Subscribe()
		assert.Equal(t, int32(1), gs.Count())
		go func() {
			defer gs.Unsubscribe()
			<-ctx.Done()
			graceful = true
		}()
		err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		assert.NoError(t, err)
		gs.Wait()

		assert.Equal(t, int32(0), gs.Count())
		assert.True(t, graceful)
	})

	t.Run("CancelFn", func(t *testing.T) {
		var graceful bool
		gs, ctx, cancel := NewContext(context.Background(), syscall.SIGINT)

		gs.Subscribe()
		assert.Equal(t, int32(1), gs.Count())
		go func() {
			defer gs.Unsubscribe()
			<-ctx.Done()
			graceful = true
		}()
		cancel()
		gs.Wait()

		assert.Equal(t, int32(0), gs.Count())
		assert.True(t, graceful)
	})
}

func Test_GracefulShutdown_Channel(t *testing.T) {
	t.Parallel()

	var graceful bool
	gs, stopCh := NewChannel(syscall.SIGINT)

	gs.Subscribe()
	assert.Equal(t, int32(1), gs.Count())
	go func() {
		defer gs.Unsubscribe()
		<-stopCh
		graceful = true
	}()
	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	assert.NoError(t, err)
	gs.Wait()

	assert.Equal(t, int32(0), gs.Count())
	assert.True(t, graceful)
}

func Test_GracefulShutdown_Negative_Count(t *testing.T) {
	t.Parallel()
	gs, _, _ := NewContext(context.Background(), syscall.SIGINT)

	gs.Unsubscribe()
	assert.Equal(t, int32(0), gs.Count())

	gs.SubscribeN(3)
	assert.Equal(t, int32(3), gs.Count())

	gs.UnsubscribeN(10)
	assert.Equal(t, int32(0), gs.Count())

	gs.UnsubscribeN(10)
	assert.Equal(t, int32(0), gs.Count())
}

func Test_GracefulShutdown_UnsubscribeFn(t *testing.T) {
	t.Parallel()
	gs, _, _ := NewContext(context.Background(), syscall.SIGINT)

	gs.Subscribe()
	assert.Equal(t, int32(1), gs.Count())

	var isDone bool
	gs.UnsubscribeFn(func() {
		shortDelay()
		isDone = true
	})
	assert.Equal(t, int32(0), gs.Count())
	assert.True(t, isDone)

	gs.UnsubscribeFn(func() {
		isDone = false
	})
	assert.Equal(t, int32(0), gs.Count())
	assert.True(t, isDone)
}

func Test_GracefulShutdown_UnsubscribeFnWithTimeout(t *testing.T) {
	t.Parallel()
	gs, _, _ := NewContext(context.Background(), syscall.SIGINT)

	gs.SubscribeN(2)
	assert.Equal(t, int32(2), gs.Count())

	var isDone bool
	gs.UnsubscribeFnWithTimeout(func() {
		shortDelay()
		isDone = true
	}, LongDelay)
	assert.Equal(t, int32(1), gs.Count())
	assert.True(t, isDone)

	isDone = false
	assert.False(t, isDone)
	gs.UnsubscribeFnWithTimeout(func() {
		longDelay()
		isDone = true
	}, ShortDelay)
	assert.Equal(t, int32(0), gs.Count())
	assert.False(t, isDone)

	gs.UnsubscribeFnWithTimeout(func() {
		isDone = true
	}, 1)
	assert.Equal(t, int32(0), gs.Count())
	assert.False(t, isDone)
}

func Test_GracefulShutdown_WaitWithTimeout(t *testing.T) {
	t.Parallel()
	gs, _, _ := NewContext(context.Background(), syscall.SIGINT)

	gs.SubscribeN(10)
	assert.Equal(t, int32(10), gs.Count())

	gs.WaitWithTimeout(ShortDelay)
	assert.Equal(t, int32(0), gs.Count())

	gs.Subscribe()
	assert.Equal(t, int32(1), gs.Count())

	go func() {
		shortDelay()
		gs.Unsubscribe()
	}()

	gs.WaitWithTimeout(LongDelay)
	assert.Equal(t, int32(0), gs.Count())
}

func shortDelay() {
	time.Sleep(ShortDelay)
}

func longDelay() {
	time.Sleep(LongDelay)
}
