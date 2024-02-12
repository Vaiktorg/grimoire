package tests

import (
	"context"
	"fmt"
	"github.com/vaiktorg/grimoire/authentity/src"
	"github.com/vaiktorg/grimoire/authentity/src/models"
	"github.com/vaiktorg/grimoire/gwt"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/names"
	"github.com/vaiktorg/grimoire/uid"
	"os"
	"testing"
)

var (
	Logger     = log.NewSimLogger("Authentity")
	ServerName = names.NewName() + "_" + string(uid.New())
	Auth       = src.NewAuthentity(&src.Config{
		Issuer: ServerName,
		GSpice: gwt.Spice{
			Salt:   []byte(uid.New()),
			Pepper: []byte(uid.New()),
		},
		Logger: Logger,
	})
	TestProfile = models.Profile{
		FirstName:   "John",
		Initial:     "E",
		LastName:    "Smith",
		PhoneNumber: "1234567890",
		Address: &models.Address{
			Addr1:   "666 Hellsing Ave.",
			Addr2:   "Bldg. 3x6 Apt. 543",
			City:    "Gehenna",
			State:   "3rd Circle",
			Country: "Inferno",
			Zip:     "00666",
		},
	}

	TestAccount = models.Account{
		Username: "nyarlathotep",
		Email:    "space-worm@elder1s.com",
		Password: "MrN00dle$123",
	}

	Token = gwt.Token{}
)

func TestMain(m *testing.M) {
	if Auth == nil {
		panic("auth is nil")
	}

	if _, err := os.Stat(ServerName); os.IsNotExist(err) {
		panic("sql db not created")
	}

	m.Run()
}

func TestAuthentityHappyPath(t *testing.T) {
	t.Run("TestRegister", func(t *testing.T) {
		err := Auth.RegisterIdentity(
			context.Background(),
			&TestProfile,
			&TestAccount,
		)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	})

	t.Run("TestLogin", func(t *testing.T) {
		tkn, err := Auth.LoginManual(
			context.Background(),
			TestAccount.Username,
			TestAccount.Password,
		)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		Token.Token = tkn.Token
		fmt.Println(Token)
	})

	t.Run("TestLoginToken", func(t *testing.T) {
		err := Auth.LoginToken(Token.Token)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	})

	t.Run("TestLogoutToken", func(t *testing.T) {
		err := Auth.LogoutToken(context.Background(), Token.Token)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	})
}
