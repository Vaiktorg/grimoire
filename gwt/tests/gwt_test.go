package tests

import (
	"bytes"
	"github.com/vaiktorg/grimoire/gwt"
	"github.com/vaiktorg/grimoire/uid"
	"testing"
	"time"
)

var mc, err = gwt.NewMultiCoder[*gwt.Resources](spice)

func TestMain(m *testing.M) {
	defer m.Run()
	if err != nil {
		panic(err)
	}
}

var TestAccount = &gwt.GWT[*gwt.Resources]{
	Header: gwt.Header{
		Issuer:    []byte("Authentity"),
		Recipient: []byte("Vaiktorg"),
		Expires:   time.Now().Add(gwt.TokenExpireTime),
	},
	Body: gwt.NewResources(uid.NewUID(gwt.FixedIDLen)),
}

var spice = gwt.Spice{
	Salt:   []byte("salt"),
	Pepper: []byte("pepper"),
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

	t.Logf("%+v", token)

	if err != nil {
		t.Errorf(err.Error())
	}
}
func TestDecodeGWT(t *testing.T) {
	if TestAccount.Token == "" {
		t.Errorf("token string is empty")
		t.FailNow()
	}

	token, err := mc.Decode(TestAccount.Token)
	if err = token.ValidateGWT(spice); err != nil {
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
