package tests

import (
	"context"
	"fmt"
	"github.com/vaiktorg/grimoire/authentity/src"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"github.com/vaiktorg/grimoire/gwt"
	"gorm.io/driver/sqlite"
	"os"
	"testing"
)

var (
	DBName      = "test_db"
	SQL         = sqlite.Open(DBName)
	Auth        = src.NewAuthentity("TestAuthentity", SQL)
	TestProfile = &entities.Profile{
		FirstName:   "John",
		Initial:     "E",
		LastName:    "Smith",
		LastName2:   "Johnson",
		PhoneNumber: "1234567890",
		Address: &entities.Address{
			Addr1:   "666 Hellsing Ave.",
			Addr2:   "Bldg. 3x6 Apt. 543",
			City:    "Gehenna",
			State:   "3rd Circle",
			Country: "Inferno",
			Zip:     "00666",
		},
	}

	TestAccount = &entities.Account{
		Username: "nyarlathotep",
		Email:    "space-worm@elder1s.com",
		Password: "MrN00dle$123",
	}

	Token           = gwt.Token{}
	Context, cancel = context.WithCancel(context.Background())
)

func TestMain(m *testing.M) {
	defer cancel()

	if Auth == nil {
		panic("auth is nil")
	}

	if _, err := os.Stat(DBName); os.IsNotExist(err) {
		panic("sql db not created")
	}

	m.Run()
}

func TestRegister(t *testing.T) {
	err := Auth.RegisterIdentity(
		Context,
		TestProfile,
		TestAccount,
	)
	var tkn *gwt.GWT[src.AuthBody]
	if err != nil {
		tkn, err = Auth.LoginManual(
			Context,
			TestAccount.Username,
			TestAccount.Password,
		)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	if tkn == nil {
		t.Error("token is nil")
		t.FailNow()
	}

	Token.Token = tkn.Token.Token
	Token.Signature = tkn.Token.Signature

	fmt.Println(Token)
}

func TestLogin(t *testing.T) {
	if Token.Token == "" {
		tkn, err := Auth.LoginManual(
			Context,
			TestAccount.Username,
			TestAccount.Password,
		)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		Token.Token = tkn.Token.Token
		Token.Signature = tkn.Token.Signature

		fmt.Println(Token)
	}
}

func TestLoginToken(t *testing.T) {
	err := Auth.LoginToken(Context, Token.Token)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func TestLogoutToken(t *testing.T) {
	err := Auth.LogoutToken(Context, Token.Token)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}
