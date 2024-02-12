package main

import (
	"fmt"
	"github.com/vaiktorg/grimoire/async"
)

func main() {
	strProm := async.AwaitHandler[string](func() (string, error) {
		return "Hello World", nil
	})

	intProm := async.AwaitHandler[int](func() (int, error) {
		return 123, nil
	})

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
