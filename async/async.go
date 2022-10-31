package main

import "fmt"

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

type AsyncStr struct {
	msg string
}
type AsyncInt struct {
	num int
}

func (a *AsyncStr) Exec() (string, error) {
	return a.msg, nil
}
func (a *AsyncInt) Exec() (int, error) {
	return a.num, nil
}

func main() {
	aStr := &AsyncStr{msg: "Hello World"}
	aInt := &AsyncInt{num: 1234567890}

	strProm := Async[string](aStr)
	intProm := Async[int](aInt)

	err := strProm.Then(func(str string) error {
		fmt.Println(str)

		return nil
	})
	if err != nil {
		println(err)
	}

	err = intProm.Then(func(num int) error {
		fmt.Println(num)

		return nil
	})
	if err != nil {
		println(err)
	}
}
