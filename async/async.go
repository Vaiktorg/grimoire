package async

// IAsync -----------------------------------------------------------------------------------------------------
// Implement this to be executed asynchronously on IAsync Operations
type IAsync[T comparable] interface {
	Exec() (T, error)
}

// Promise ----------------------------------------------------------------------------------------------------
type Promise[T comparable] interface {
	Then(func(T) error) error
}

// Async ----------------------------------------------------------------------------------------------------
func Async[T comparable](await IAsync[T]) Promise[T] {
	result := PromiseRes[T]{Value: make(chan T), Err: make(chan error)}

	go func(async IAsync[T], res PromiseRes[T]) {
		defer close(res.Value)
		defer close(res.Err)

		v, e := async.Exec()
		if e != nil {
			res.Err <- e
			return
		}

		res.Value <- v
	}(await, result)

	return result
}

type PromiseRes[T comparable] struct {
	Value chan T
	Err   chan error
}

func (r PromiseRes[T]) Then(handler func(T) error) error {
	select {
	case v := <-r.Value:
		return handler(v)
	case e := <-r.Err:
		return e
	}
}

// AsyncOp ----------------------------------------------------------------------------------------------------

type AsyncOp[T comparable] struct {
	Val T
}

func (a *AsyncOp[T]) Exec() (T, error) {
	return a.Val, nil
}
