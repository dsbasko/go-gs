# Go Graceful Shutdown Library
Package gogs provides a mechanism for managing graceful shutdowns in Go applications.

It defines an interface, GracefulShutdowner, with methods for subscribing and 
unsubscribing to shutdown events, and waiting for all events to complete. It also provides 
a concrete implementation of this interface, GracefulShutdown.

GracefulShutdown uses a sync.WaitGroup to wait for all active shutdown events to complete, 
and an atomic.Int32 to keep track of the count of active events. The package also provides 
functions for creating a new context or channel that can be used to signal shutdown events.

Need go version 1.19 or later.

### Installation
To install the library, use the go get command:
```bash
go get github.com/dsbasko/go-gs
```

## Usage
Here is an example of how to use the library:

**Code:**

```go
package main

import (
	"context"
	"log"
	"syscall"
	"time"

	gogs "github.com/dsbasko/go-gs"
)

func main() {
	// Create a new context for GS.
	gs, ctx, cancel := gogs.NewContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Subscribe to the GS and someFn in a goroutine.
	gs.Subscribe()
	go someFn(ctx, gs)

	// Subscribe to the GS with a count of 5 and start 5 workers in a goroutine.
	gs.SubscribeN(5)
	go someWorker(ctx, gs, 5)

	// Adding a delay to complete multiple tasks from a worker.
	time.Sleep(time.Second)
	cancel()

	// Wait for all subscribed functions to complete.
	gs.Wait()
	log.Println("main: graceful shutdown is done!")
}

// someFn is a function that, after receiving a cancellation signal, waits 1 second and
// unsubscribes from GS.
func someFn(ctx context.Context, gs gogs.GracefulShutdowner) {
	<-ctx.Done()

	gs.UnsubscribeFn(func() {
		time.Sleep(1 * time.Second)
		log.Println("someFn: graceful shutdown!")
	})
}

// someWorker is a function that starts a number of workers equal to count and logs a
// message every 450 milliseconds. After receiving a cancellation signal, it unsubscribes
// from GS and logs a message after a delay equal to the worker number in seconds.
func someWorker(ctx context.Context, gs gogs.GracefulShutdowner, count int) {
	t := time.NewTicker(450 * time.Millisecond)

	for i := 0; i < count; i++ {
		go func(i int) {
			// Loop until the context is done.
			for {
				select {
				case <-ctx.Done():
					// Unsubscribe from the GS.
					gs.UnsubscribeFn(func() {
						time.Sleep(time.Duration(i) * time.Second)
						log.Printf("worker %v: graceful shutdown!\n", i+1)
					})
					return
				case <-t.C:
					// Log that a job is done.
					log.Printf("job is done...")
				}
			}
		}(i)
	}
}
```

**Logic**:  

1. _In the main function, a new context for gogs is created using the NewContext function. 
    This context will monitor the interrupt and shutdown signals;_  
2. _The Subscribe function is then called to subscribe to gogs events;_  
3. _The someFn function runs in a separate goroutine. It waits for a cancellation signal 
    from the context, after which it unsubscribes from gogs, performing a function that 
    waits one second and displays a message about a smooth shutdown;_ 
4. _The subscribe function is called with the argument 5, which means subscribing to gogs 
    with the number of workers equal to 5;_
5. _The someWorker function also runs in a separate plugin. It creates a specified number 
    of workers, each of which outputs a message every 450 milliseconds in a loop. When a 
    cancellation signal is received from the context, the worker unsubscribes from gogs, 
    performing a function that waits for a time equal to the worker's number in seconds 
    and outputs a message about a smooth shutdown;_
6. _A one-second delay is added to the main function so that workers can complete 
    multiple tasks;_
7. _An interrupt signal is then sent, which causes the context to be canceled and the 
    smooth shutdown process to begin;_
8. _The Wait function is called to wait for all signed functions to complete;_
9. _At the end, a message is displayed stating that a smooth shutdown has been completed._

**Output**:

```
2006/02/01 00:00:00 job is done...
2006/02/01 00:00:01 job is done...
2006/02/01 00:00:01 worker 1: graceful shutdown!
2006/02/01 00:00:02 worker 2: graceful shutdown!
2006/02/01 00:00:02 someFn: graceful shutdown!
2006/02/01 00:00:03 worker 3: graceful shutdown!
2006/02/01 00:00:04 worker 4: graceful shutdown!
2006/02/01 00:00:05 worker 5: graceful shutdown!
2006/02/01 00:00:05 main: graceful shutdown is done!
```

## Constructors
```go
// Creates a new context for graceful shutdown and returns a new GracefulShutdowner, the new context, and a cancel function.
gs, ctx, cancel := gogs.NewContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

// Creates a new channel for graceful shutdown and returns a new GracefulShutdowner and the new channel.
gs, ch := gogs.NewChannel(syscall.SIGINT, syscall.SIGTERM)
```

## Methods

```go
// Increments the count of active shutdown events by one.
gs.Subscribe()

// Increments the count of active shutdown events by the specified count.
gs.SubscribeN(count int32)

// Decrements the count of active shutdown events by one.
gs.Unsubscribe()

// Decrements the count of active shutdown events by the specified count.
gs.UnsubscribeN(count int32)

// Executes the provided function and unsubscribes immediately after the
// function execution completes.
gs.UnsubscribeFn(cleanFn func())

// Executes the provided function and unsubscribes after the specified duration. If the 
// function execution completes before the timeout, it unsubscribes immediately.
gs.UnsubscribeFnWithTimeout(cleanFn func(), duration time.Duration)

// Returns the current count of active shutdown events.
gs.Count() int32

// Blocks until all active shutdown events have completed.
gs.Wait()

// Blocks until all active shutdown events have completed or the specified duration has 
// elapsed. If the duration elapses before all events have completed, it unsubscribes 
// from all remaining events.
gs.WaitWithTimeout(duration time.Duration)
```

<br>

---

If you enjoyed this project, I would appreciate it if you could give it a star! If you notice any problems or have any suggestions for improvement, please feel free to create a new issue. Your feedback means a lot to me!

❤️ [Dmitriy Basenko](https://github.com/dsbasko)