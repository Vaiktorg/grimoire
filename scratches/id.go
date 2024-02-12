package main

import (
	"github.com/vaiktorg/grimoire/uid"
	"github.com/vaiktorg/grimoire/util"
)

func main() {
	id := uid.NewUID(4096)

	mc, err := util.NewMultiCoder[string]()
	if err != nil {
		panic(err)
	}

	err = mc.EncodeSave("Global.ID", string(id), util.EncodeJson)
	if err != nil {
		panic(err)
	}
}
