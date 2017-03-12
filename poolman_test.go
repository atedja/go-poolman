/*
Copyright 2017 Albert Tedja

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package poolman

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func slowTask(wg interface{}) {
	time.Sleep(1 * time.Second)
	wg.(*sync.WaitGroup).Done()
}

func fastTask(wg interface{}) {
	time.Sleep(100 * time.Millisecond)
	wg.(*sync.WaitGroup).Done()
}

func TestDefaultPoolmanNotNil(t *testing.T) {
	assert.NotNil(t, Default)
}

func TestAddTasksSlow(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)
	Default.Resize(2)

	Default.AddTask(slowTask, &wg)
	Default.AddTask(slowTask, &wg)
	wg.Wait()
	Default.Close()
}

func TestAddTasksMany(t *testing.T) {
	var wg sync.WaitGroup
	Default.Resize(2)
	wg.Add(20)
	for i := 0; i < 20; i++ {
		Default.AddTask(fastTask, &wg)
	}
	wg.Wait()
	Default.Close()
}

func TestResize(t *testing.T) {
	var wg sync.WaitGroup

	Default.Resize(2)
	wg.Add(10)
	for i := 0; i < 10; i++ {
		Default.AddTask(fastTask, &wg)
	}
	wg.Wait()

	Default.Resize(1)
	wg.Add(10)
	for i := 0; i < 10; i++ {
		Default.AddTask(fastTask, &wg)
	}
	wg.Wait()

	Default.Resize(8)
	wg.Add(8)
	for i := 0; i < 8; i++ {
		Default.AddTask(fastTask, &wg)
	}
	wg.Wait()

	Default.Close()
}

func TestCancellableTask(t *testing.T) {
	Default.Resize(1)
	done := make(chan bool)
	ctx, cancelFn := context.WithCancel(context.Background())
	Default.AddTask(func(args ...interface{}) {
		ctx := args[0].(context.Context)
		select {
		case <-time.After(2 * time.Second):
			args[1].(chan bool) <- true
		case <-ctx.Done():
			return
		}
	}, ctx, done)

	select {
	case <-time.After(1 * time.Second):
		cancelFn()
	case <-done:
		assert.Equal(t, 0, 1, "This should not be executed")
	}
	Default.Close()
}

func TestTimeoutTask(t *testing.T) {
	Default.Resize(1)
	done := make(chan bool)
	ctx, cancelFn := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancelFn()

	Default.AddTask(func(args ...interface{}) {
		ctx := args[0].(context.Context)
		done := args[1].(chan bool)
		select {
		case <-time.After(2 * time.Second):
			done <- true
		case <-ctx.Done():
			done <- false
			return
		}
	}, ctx, done)

	result := <-done
	assert.Equal(t, false, result)
	Default.Close()
}
