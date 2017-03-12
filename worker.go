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

type worker struct {
	stopped chan bool
}

func (self *worker) run(incoming <-chan *task) {
OuterLoop:
	for {
		select {
		case tsk := <-incoming:
			// Unfortunately, we can't use reflection here
			if len(tsk.Args) > 1 {
				tsk.Fn.(func(...interface{}))(tsk.Args...)
			} else if len(tsk.Args) == 1 {
				tsk.Fn.(func(interface{}))(tsk.Args[0])
			} else {
				tsk.Fn.(func())()
			}

		case <-self.stopped:
			break OuterLoop
		}
	}
}

func (self *worker) stop() {
	self.stopped <- true
}
