package tests

import (
	"bytes"
	"github.com/vaiktorg/grimoire/gwt"
	"github.com/vaiktorg/grimoire/uid"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	defer m.Run()
}

var TestAccount = &gwt.GWT[*gwt.Resources]{
	Header: gwt.Header{
		Issuer:    []byte("Authentity"),
		Recipient: []byte("Vaiktorg"),
		Expires:   time.Now().Add(gwt.TokenExpireTime),
	},
	Body: gwt.NewResources(uid.New()),
}

func TestEncodeGWT(t *testing.T) {
	token, err := mc.Encode(TestAccount)
	if token.Token == "" {
		t.Errorf("token string is empty")
		t.FailNow()
	}

	if token.Signature == "" {
		t.Errorf("token signature is empty")
		t.FailNow()
	}

	TestAccount.Token = token.Token

	if err != nil {
		t.Errorf(err.Error())
	}
}
func TestDecodeGWT(t *testing.T) {
	var mc, err = gwt.NewMultiCoder[*gwt.Resources]()
	if err != nil {
		panic(err)
	}

	if TestAccount.Token == "" {
		t.Errorf("token string is empty")
		t.FailNow()
	}

	token, err := mc.Decode(TestAccount.Token)
	if err = gwt.ValidateGWT(token); err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}

	if !bytes.Equal(token.Header.Recipient, TestAccount.Header.Recipient) {
		t.Errorf("mismatch recipient")
	}
	if token.Header.Expires.Compare(TestAccount.Header.Expires) != 0 {
		t.Errorf("mismatch expire date")
	}

	if !bytes.Equal(token.Header.Issuer, TestAccount.Header.Issuer) {
		t.Errorf("mismatch issuer")
	}
	if token.Token != TestAccount.Token {
		t.Errorf("mismatch token")
	}

	if err != nil {
		t.Errorf(err.Error())
	}
}
