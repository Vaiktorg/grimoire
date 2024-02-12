package tests

import (
	"github.com/vaiktorg/grimoire/gwt"
	"github.com/vaiktorg/grimoire/uid"
	"reflect"
	"testing"
)

func TestSerializationDeserialization(t *testing.T) {
	original := gwt.NewResources(uid.New())

	res1 := gwt.NewResource(gwt.DataManagement, gwt.DefaultRoles[gwt.Dev])
	original.AddResource(res1)

	res2 := gwt.NewResource(gwt.Network, gwt.DefaultRoles[gwt.Dev])
	original.AddResource(res2)

	// Serialize
	serialized := original.Serialize()

	// Deserialize
	resources := &gwt.Resources{}
	err := resources.Deserialize(serialized)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
		return
	}

	// Compare original and deserialized
	if !reflect.DeepEqual(original.Serialize(), resources.Serialize()) {
		t.Errorf("Original: %+v\nDeserialized: %+v\n", original, resources)
		return
	}
}
