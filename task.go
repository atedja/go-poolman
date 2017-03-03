package poolman

type task struct {
	Fn   interface{}
	Args []interface{}
}
