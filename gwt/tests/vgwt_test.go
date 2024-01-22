package tests

import (
	"bytes"
	"fmt"
	"github.com/vaiktorg/grimoire/gwt"
	"testing"
	"time"
)

var mc, _ = gwt.NewMultiCoder[*gwt.Resources]()

func TestVGWT(t *testing.T) {
	// Resources
	original := gwt.NewResources([]byte("This is the UserID"))

	res1 := gwt.NewResource(gwt.NetworkDatabaseAPI, gwt.DefaultRoles[gwt.Dev])
	original.AddResource(res1)

	res2 := gwt.NewResource(gwt.DevToolsCICD, gwt.DefaultRoles[gwt.Dev])
	original.AddResource(res2)

	// GoWebToken
	gwTok := &gwt.GWT[*gwt.Resources]{
		Header: gwt.Header{
			Issuer:    []byte("Authentity"),
			Recipient: []byte("VKTRG"),
			Expires:   time.Now().UTC().Add(time.Minute * 15),
		},
		Body: original,
	}

	tok, err := mc.Encode(gwTok)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	gwt.SetVTokenConfig(gwt.SmallConfig)

	var originalHash []byte
	if originalHash, err = tok.CreateTokenCard(); err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	var decodedHash []byte
	decodedHash, err = tok.ReadTokenCard("id_card.png")
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	fmt.Println(originalHash)
	fmt.Println(decodedHash)

	fmt.Println(gwt.XORText(originalHash, gwt.HashKey))
	fmt.Println(gwt.XORText(decodedHash, gwt.HashKey))
	fmt.Println(tok.Signature)

	if !bytes.Equal(decodedHash, originalHash) {
		t.Error("decodedHash and token signature do not match")
		t.FailNow()
	}

}
