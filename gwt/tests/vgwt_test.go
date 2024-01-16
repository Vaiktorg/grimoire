package tests

import (
	"bytes"
	"fmt"
	"github.com/vaiktorg/grimoire/gwt"
	"github.com/vaiktorg/grimoire/uid"
	"testing"
)

var vh *gwt.VisualHash[*gwt.Resources]

func TestVGWT(t *testing.T) {
	vh, err = gwt.NewVisualHash[*gwt.Resources]()
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	original := gwt.NewResources(uid.NewUID(gwt.FixedIDLen))

	res1 := gwt.NewResource(gwt.NetworkDatabaseAPI, gwt.DefaultRoles[gwt.Dev])
	original.AddResource(res1)

	res2 := gwt.NewResource(gwt.DevToolsCICD, gwt.DefaultRoles[gwt.Dev])
	original.AddResource(res2)

	var originalHash []byte
	if originalHash, err = vh.CreateHashCard(original); err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	var hash []byte
	hash, err = vh.ReadHashCard("id_card.png")
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	fmt.Println(originalHash)
	fmt.Println(hash)

	if !bytes.Equal(hash, originalHash) {
		t.Error("hash and token signature do not match")
		t.FailNow()
	}
}
func padByteArray(data []byte, length int) []byte {
	if len(data) >= length {
		return data[:length] // Truncate if longer
	}
	padded := make([]byte, length)
	copy(padded, data)
	// The rest of padded will be zero-valued
	return padded
}
