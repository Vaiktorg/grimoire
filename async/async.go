package async

// IAsync -----------------------------------------------------------------------------------------------------
// Implement this to be executed asynchronously on IAsync Operations
type IAsync[T any] interface {
	Exec() (T, error)
}

// IPromise ----------------------------------------------------------------------------------------------------
type IPromise[T any] interface {
	Then(func(T) error) error
}

func AwaitHandler[T any](await func() (T, error)) IPromise[T] {
	result := Promise[T]{Value: make(chan T), Err: make(chan error)}

	go func(async func() (T, error), res Promise[T]) {
		defer close(res.Value)
		defer close(res.Err)

		v, e := async()
		if e != nil {
			res.Err <- e
			return
		}

		res.Value <- v
	}(await, result)

	return result
}

// Await ----------------------------------------------------------------------------------------------------
func Await[T any](await IAsync[T]) IPromise[T] {
	result := Promise[T]{Value: make(chan T), Err: make(chan error)}

	go func(async IAsync[T], res Promise[T]) {
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

type Promise[T any] struct {
	Value chan T
	Err   chan error
}

func (r Promise[T]) Then(handler func(T) error) error {
	select {
	case v := <-r.Value:
		return handler(v)
	case e := <-r.Err:
		return e
	}
}
