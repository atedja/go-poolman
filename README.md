# Poolman

Poolman is a goroutine pool manager to manage your asynchronous tasks.
It is a simple library meant to control the number of background workers.

### Quick Example

```go
package main

import (
  "fmt"
  "sync"
  "github.com/atedja/go-poolman"
)

func main() {
  var wg sync.WaitGroup
  wg.Add(2)

  // No parameter
  poolman.Default.AddTask(func() {
    fmt.Println("Executing a long running task...")
    time.Sleep(10 * time.Second)
    fmt.Println("Done")
    wg.Done()
  })

  // With parameters
  poolman.Default.AddTask(func(params ...interface{}) {
    job := params[0].(string)
    jv := params[1].(int)
    fmt.Println("Another long running task:", job)
    time.Sleep(jv * time.Second)
    fmt.Println("Other task done")
    wg.Done()
  }, "foo", 5)

  wg.Wait()
}
```


### Basic Usage

#### Creating a Poolman

    var pm *poolman.Poolman
    pm = poolman.New(8, 16) // Create 8 background workers and a queue of size 16

Poolman comes with a basic Poolman `Default`, as shown in the example above. The default Poolman has numbers of workers equals to the value returned by `runtime.NumCPU()`.

#### Add an asynchronous task

    pm := poolman.New(4, 16)
    poolman.Default.AddTask(func(params ...interface{}) {
      job := params[0].(string)
      jv := params[1].(int)
      fmt.Println("Long running task:", job)
      time.Sleep(jv * time.Second)
      fmt.Println("It's Done!")
      wg.Done()
    }, "foo", 5)

#### Resize Number of Workers

    pm := poolman.New(4, 16)
    pm.Resize(8)

If new size is larger, Poolman will allocate more workers leaving the currently running workers, and the queue, untouched.
If new size is smaller, Poolman will instruct the excess old workers to stop gracefully.

#### Close Poolman

Close Poolman to release all workers

    pm.Close()


### FAQ & Best Practices

#### Preventing Deadlocks

Poolman is made intentionally simple.  While Poolman itself is robust, it cannot prevent deadlocked goroutines from crashing your system.
While it is easy to make Poolman time out the deadlocked workers, there is no known way to kill a deadlocked goroutine.

Make sure that your asynchronous tasks cannot deadlock each other.

#### Cancellation, Deadline, and Timing out Background Tasks

If you need a way to cancel an asynchronous task that has been sent to Poolman, or timing it out, it is highly recommended that you use
`context.Context`.  Poolman itself does not come with this feature because there is no way to cleanly implement timeout or job cancellation
without potentially exposing yourself to memory leaks and hidden bugs.

The following is an example of how to implement an asynchronous task with timeout.

    pm := poolman.New(4, 16)
    done := make(chan bool)
    ctx, cancelFn := context.WithTimeout(context.Background(), 5 * time.Second)
    defer cancelFn()
    
    Default.AddTask(func(params ...interface{}) {
      ctx := params[0].(context.Context)
      done := params[1].(chan bool)
      select {
      case <-time.After(10 * time.Second):
        fmt.Println("Job's done!")
        done <- true
      case <-ctx.Done():
        fmt.Println("Timeout!")
        done <- false
        return
      }
    }, ctx, done)
    
    result := <-done

#### I want synchronous tasks

Then just execute your functions directly. You do not need Poolman for that.

#### How do I get the return value from the asynchronous tasks?

Use channels, as demonstrated in the above example.
  

#### [Full API Documentation](https://godoc.org/github.com/atedja/go-poolman)
