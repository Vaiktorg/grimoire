package main

import (
	"fmt"
	"github.com/vaiktorg/grimoire/async"
)

func main() {
	aStr := &async.AsyncOp[string]{Val: "Hello World"}
	aInt := &async.AsyncOp[int]{Val: 1234567890}

	strProm := async.Async[string](aStr)
	intProm := async.Async[int](aInt)

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
