package tests

import (
	"github.com/vaiktorg/grimoire/gwt"
	"log"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	defer m.Run()
}

type Account struct {
	gwt.Value
	Username string
	Password string
}

var ResTok gwt.Token
var TestAccount = Account{
	Value: gwt.Value{
		Issuer:    "Vaiktorg",
		Recipient: "Test",
		Timestamp: time.Now(),
	},
	Username: "username-test-123",
	Password: "password-test-123",
}

func TestEncodeGWT(t *testing.T) {
	enc := gwt.NewEncoder[Account](gwt.Spice{
		Salt: []byte("salt"),
	})

	err := enc.Encode(TestAccount, func(token gwt.Token) error {
		if token.Token == "" {
			t.Errorf("token string is empty")
			t.FailNow()
		}

		if token.Signature == nil {
			t.Errorf("token signature is empty")
			t.FailNow()
		}

		ResTok = token
		log.Println(token)

		return nil
	})
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestDecodeGWT(t *testing.T) {
	dec := gwt.NewDecoder[Account](gwt.Spice{
		Salt: []byte("salt"),
	})

	if ResTok.Token == "" {
		t.Errorf("token string is empty")
		t.FailNow()
	}

	if ResTok.Signature == nil {
		t.Errorf("token string is empty")
		t.FailNow()
	}

	err := dec.Decode(ResTok, func(account Account) error {
		if !dec.ValidateSignature(account, ResTok.Signature) {
			t.Errorf("invalid signature")
		}

		if account.Username != TestAccount.Username ||
			account.Password != TestAccount.Password ||
			account.Value.Issuer != TestAccount.Value.Issuer ||
			account.Value.Recipient != TestAccount.Value.Recipient ||
			account.Value.Timestamp != account.Value.Timestamp {

			t.Errorf("decoding failed, field mismatch")
			t.FailNow()
		}

		return nil
	})
	if err != nil {
		t.Errorf(err.Error())
	}
}
