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
