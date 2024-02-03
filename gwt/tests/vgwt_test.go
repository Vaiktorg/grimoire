package tests

import (
	"bytes"
	"fmt"
	"github.com/vaiktorg/grimoire/gwt"
	"github.com/vaiktorg/grimoire/gwt/vhash"
	"image"
	"testing"
	"time"
)

var mc, _ = gwt.NewMultiCoder[*gwt.Resources]()

func TestVGWT(t *testing.T) {
	// Resources
	original := gwt.NewResources("This is the UserID")

	res1 := gwt.NewResource(gwt.Network, gwt.DefaultRoles[gwt.Dev])
	original.AddResource(res1)

	res2 := gwt.NewResource(gwt.DataManagement, gwt.DefaultRoles[gwt.Dev])
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

	//var maskHash []byte
	var mask *image.RGBA
	_, mask, err = vhash.CreateTokenCard([]byte(tok.Signature))
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	var decodedHash []byte
	decodedHash = vhash.ReadTokenCard(vhash.GridConfig.ExportPath, mask)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	fmt.Println(tok.Signature)
	fmt.Println(decodedHash)
	fmt.Println(tok.Signature)

	if !bytes.Equal(decodedHash, []byte(tok.Signature)) {
		t.Log(decodedHash)
		t.Log(tok.Signature)
		t.Error("decodedHash and originalHash do not match")
		t.FailNow()
		return
	}
}
