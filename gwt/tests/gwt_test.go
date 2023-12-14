package tests

import (
	"bytes"
	"github.com/vaiktorg/grimoire/authentity/src"
	"github.com/vaiktorg/grimoire/gwt"
	"log"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	defer m.Run()
}

var TestAccount = gwt.GWT[src.AuthBody]{
	Header: gwt.Header{
		Issuer:    "Authentity",
		Recipient: "Vaiktorg",
		Expires:   time.Now(),
	},
	Body: src.AuthBody{
		Permission: 0,
	},
}

func TestEncodeGWT(t *testing.T) {
	enc := gwt.NewEncoder[*gwt.GWT[src.AuthBody]](gwt.Spice{
		Salt:   []byte("salt"),
		Pepper: []byte("pepper"),
	})

	token, err := enc.Encode(&TestAccount)
	if token.Token == "" {
		t.Errorf("token string is empty")
		t.FailNow()
	}

	if token.Signature == nil {
		t.Errorf("token signature is empty")
		t.FailNow()
	}

	TestAccount.Token.Token = token.Token
	TestAccount.Token.Signature = token.Signature

	t.Logf("%+v", token)

	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestDecodeGWT(t *testing.T) {
	spice := gwt.Spice{
		Salt:   []byte("salt"),
		Pepper: []byte("pepper"),
	}

	dec := gwt.NewDecoder[*gwt.GWT[src.AuthBody]](spice)

	if TestAccount.Token.Token == "" {
		t.Errorf("token string is empty")
		t.FailNow()
	}

	token, err := dec.Decode(TestAccount.Token.Token)
	if err = token.ValidateGWT(nil); err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}

	if !bytes.Equal(token.Token.Signature, TestAccount.Token.Signature) {
		t.Error("non matching signatures")
		return
	}

	if token.Header.Recipient != TestAccount.Header.Recipient {
		t.Errorf("mismatch recipient")
	}
	if token.Header.Expires.Compare(TestAccount.Header.Expires) != 0 {
		t.Errorf("mismatch expire date")
	}

	if token.Header.Issuer != TestAccount.Header.Issuer {
		t.Errorf("mismatch issuer")
	}
	if token.Token.Token != TestAccount.Token.Token {
		t.Errorf("mismatch token.token")
	}

	if !bytes.Equal(token.Token.Signature, TestAccount.Token.Signature) {
		t.Errorf("mismatch token.signature")
		t.FailNow()
	}

	log.Printf("%+v", token)

	if err != nil {
		t.Errorf(err.Error())
	}
}
