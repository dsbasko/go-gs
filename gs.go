// Package gogs provides a mechanism for managing graceful shutdowns in Go applications.
//
// It defines an interface, GracefulShutdowner, with methods for subscribing and
// unsubscribing to shutdown events, and waiting for all events to complete. It also
// provides a concrete implementation of this interface, GracefulShutdown.
//
// GracefulShutdown uses a sync.WaitGroup to wait for all active shutdown events to
// complete, and an atomic.Int32 to keep track of the count of active events. The package
// also provides functions for creating a new context or channel that can be used to
// signal shutdown events.
package gogs

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

// GracefulShutdowner is an interface that provides methods for managing graceful
// shutdowns. It allows subscribing and unsubscribing to shutdown events, and waiting for
// all events to complete.
type GracefulShutdowner interface {
	// Subscribe increments the count of active shutdown events by one.
	Subscribe()

	// SubscribeN increments the count of active shutdown events by the specified count.
	SubscribeN(count int32)

	// Unsubscribe decrements the count of active shutdown events by one.
	Unsubscribe()

	// UnsubscribeN decrements the count of active shutdown events by the specified count.
	UnsubscribeN(count int32)

	// UnsubscribeFn executes the provided function and unsubscribes immediately after the
	// function execution completes.
	UnsubscribeFn(cleanFn func())

	// UnsubscribeFnWithTimeout executes the provided function and unsubscribes after the
	// specified duration. If the function execution completes before the timeout, it
	// unsubscribes immediately.
	UnsubscribeFnWithTimeout(cleanFn func(), duration time.Duration)

	// Count returns the current count of active shutdown events.
	Count() int32

	// Wait blocks until all active shutdown events have completed.
	Wait()

	// WaitWithTimeout blocks until all active shutdown events have completed or the
	// specified duration has elapsed. If the duration elapses before all events have
	// completed, it unsubscribes from all remaining events.
	WaitWithTimeout(duration time.Duration)
}

// GracefulShutdown is a struct that implements the GracefulShutdowner interface.
// It provides a mechanism for managing graceful shutdowns in Go applications.
// It uses a sync.WaitGroup to wait for all active shutdown events to complete,
// and an atomic.Int32 to keep track of the count of active events.
type GracefulShutdown struct {
	// wg is a WaitGroup that is used to wait for all active shutdown events to complete.
	wg sync.WaitGroup

	// list is an atomic integer that keeps track of the count of active shutdown events.
	list atomic.Int32
}

// NewContext is a function that creates a new context and a GracefulShutdowner instance.
// It takes a parent context and a variadic parameter of os.Signal as arguments.
// The function uses the signal.NotifyContext function to register the provided signals to
// the created context.
//
//	gs, ctx, cancel := NewContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
//
// This example creates a new context that will be canceled when an interrupt or
// termination signal is received. It also returns a GracefulShutdowner instance that can
// be used to manage graceful shutdowns in the application.
func NewContext(parentCtx context.Context, signals ...os.Signal) (GracefulShutdowner, context.Context, context.CancelFunc) {
	ctx, cancel := signal.NotifyContext(parentCtx, signals...)
	return &GracefulShutdown{}, ctx, cancel
}

// NewChannel is a function that creates a new channel and a GracefulShutdowner instance.
// It takes a variadic parameter of os.Signal as arguments. The function uses the
// signal.Notify function to register the provided signals to the created channel.
//
//	gs, stopCh := NewChannel(syscall.SIGINT, syscall.SIGTERM)
//
// This example creates a new channel that will receive an interrupt or termination
// signal. It also returns a GracefulShutdowner instance that can be used to manage
// graceful shutdowns in the application.
func NewChannel(signals ...os.Signal) (GracefulShutdowner, chan os.Signal) {
	stopCh := make(chan os.Signal, 2)
	signal.Notify(stopCh, signals...)
	return &GracefulShutdown{}, stopCh
}

// Subscribe is a method of the GracefulShutdown struct. It increments the count of active
// shutdown events by one.
func (gs *GracefulShutdown) Subscribe() {
	gs.list.Add(1)
	gs.wg.Add(1)
}

// SubscribeN is a method of the GracefulShutdown struct. It increments the count of
// active shutdown events by the specified count.
func (gs *GracefulShutdown) SubscribeN(count int32) {
	gs.list.Add(count)
	gs.wg.Add(int(count))
}

// Unsubscribe is a method of the GracefulShutdown struct. It decrements the count of
// active shutdown events by one.
func (gs *GracefulShutdown) Unsubscribe() {
	if gs.list.Load() == 0 {
		return
	}
	gs.list.Add(-1)
	gs.wg.Done()
}

// UnsubscribeN is a method of the GracefulShutdown struct. It decrements the count of
// active shutdown events by the specified count.
func (gs *GracefulShutdown) UnsubscribeN(count int32) {
	list := gs.list.Load()
	if list == 0 {
		return
	}

	if list < count {
		count = list
	}

	gs.list.Add(count * -1)
	for i := int32(0); i < count; i++ {
		gs.wg.Done()
	}
}

// UnsubscribeFn is a method of the GracefulShutdown struct. It executes the provided
// function and unsubscribes immediately after the function execution completes.
func (gs *GracefulShutdown) UnsubscribeFn(cleanFn func()) {
	if gs.list.Load() == 0 {
		return
	}

	defer gs.Unsubscribe()
	cleanFn()
}

// UnsubscribeFnWithTimeout is a method of the GracefulShutdown struct. It executes the
// provided function and unsubscribes after the specified duration. If the function
// execution completes before the timeout, it unsubscribes immediately.
func (gs *GracefulShutdown) UnsubscribeFnWithTimeout(
	cleanFn func(),
	duration time.Duration,
) {
	if gs.list.Load() == 0 {
		return
	}

	defer gs.Unsubscribe()
	doneCh := make(chan struct{})

	t := time.NewTimer(duration)

	go func() {
		cleanFn()
		close(doneCh)
	}()

	select {
	case <-t.C:
		return
	case <-doneCh:
		return
	}
}

// Count is a method of the GracefulShutdown struct. It returns the current count of
// active shutdown events.
func (gs *GracefulShutdown) Count() int32 {
	return gs.list.Load()
}

// Wait is a method of the GracefulShutdown struct. It blocks until all active shutdown
// events have completed.
func (gs *GracefulShutdown) Wait() {
	gs.wg.Wait()
}

// WaitWithTimeout is a method of the GracefulShutdown struct. It blocks until all active
// shutdown events have completed or the specified duration has elapsed. If the duration
// elapses before all events have completed, it unsubscribes from all remaining events.
func (gs *GracefulShutdown) WaitWithTimeout(duration time.Duration) {
	timer := time.NewTimer(duration)
	doneCh := make(chan struct{})
	defer func() {
		<-doneCh
	}()

	go func() {
		gs.Wait()
		close(doneCh)
	}()

	select {
	case <-timer.C:
		gs.UnsubscribeN(gs.Count())
		return
	case <-doneCh:
		return
	}
}
