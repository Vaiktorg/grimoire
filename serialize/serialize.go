package main

import (
	"fmt"
	"github.com/vaiktorg/grimoire/gwt"
	"github.com/vaiktorg/grimoire/uid"
	"github.com/vaiktorg/grimoire/util"
	"time"
)

func main() {
	token := &gwt.GWT[*gwt.Resources]{
		Header: gwt.Header{
			Issuer:    []byte("me"),
			Recipient: []byte("you"),
			Expires:   time.Now(),
		},
		Body:  gwt.NewResources(uid.New()),
		Token: string(uid.NewUID(64)),
	}

	mc, err := util.NewMultiCoder[*gwt.GWT[*gwt.Resources]]()
	if err != nil {
		panic(err)
	}

	encoded, err := mc.EncodeEncrypt(token, util.EncodeGob)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(encoded))

	// ==========
	decode, err := mc.DecodeDecrypt(encoded, util.DecodeGob)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", decode)
}
