package src

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/vaiktorg/grimoire/authentity/internal"
	"github.com/vaiktorg/grimoire/authentity/src/models"
	"github.com/vaiktorg/grimoire/gwt"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/uid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type Config struct {
	Issuer string
	GSpice gwt.Spice
	Logger log.ILogger
}

type Authentity struct {
	issuer   []byte
	spice    gwt.Spice
	l        log.ILogger
	mux      *http.ServeMux
	provider *DataProvider
	mc       *gwt.MultiCoder[*gwt.Resources]
}

const issuerName = "Authenitity"

func NewAuthentity(config *Config) *Authentity {
	db, err := gorm.Open(sqlite.Open(issuerName+".db"), nil)
	if err != nil {
		panic(err)
	}

	spice := gwt.Spice{
		Salt:   []byte(uid.NewUID(8)),
		Pepper: []byte(uid.NewUID(8)),
	}

	mc, err := gwt.NewMultiCoder[*gwt.Resources](&spice)
	if err != nil {
		panic(err)
	}

	config.Logger.TRACE("Authentity entity " + issuerName + " is running")
	auth := &Authentity{
		issuer:   []byte(issuerName),
		provider: NewDataProvider(db),
		spice:    spice,
		l:        config.Logger,

		mux: http.NewServeMux(),
		mc:  mc,
	}

	if err = auth.Migrate(); err != nil && !errors.Is(err, AlreadyExistError) {
		panic(err)
	}

	auth.registerMux()

	return auth
}

func (a *Authentity) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	internal.SecurityMiddleware(a.mux).ServeHTTP(w, r)
}

func (a *Authentity) RegisterIdentity(pCtx context.Context, prof *models.Profile, acc *models.Account) error {
	ctx, cancel := context.WithTimeout(pCtx, time.Minute)
	defer cancel()

	if _, err := a.provider.AccountsService.FindAccountByEmail(ctx, acc.Email); err == nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("email already being used for email: " + acc.Email)
	}
	if _, err := a.provider.AccountsService.FindAccountByUsername(ctx, acc.Username); err == nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("username already being used for user: " + acc.Username)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(acc.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("could not create hashed password")
	}

	id := uid.New()
	tok := &gwt.GWT[*gwt.Resources]{
		Header: gwt.Header{
			Issuer:    a.issuer,
			Recipient: []byte(acc.Username),
			Expires:   time.Now().Add(gwt.TokenExpireTime),
		},
		Body: gwt.NewResources(id), // TODO: Mod Request for AssignResource
	}

	t, err := a.mc.Encode(tok)
	if err != nil {
		return err
	}

	identity := &models.Identity{
		ID:      string(id),
		Profile: prof,
		Account: &models.Account{
			Username:  acc.Username,
			Email:     acc.Email,
			Password:  string(hashedPassword),
			Signature: t.Signature,
		},
		Resources: tok.Body,
	}

	return a.provider.IdentityService.Persist(ctx, identity)
}

func (a *Authentity) LoginManual(pCtx context.Context, identifier, password string) (*gwt.GWT[*gwt.Resources], error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Minute)
	defer cancel()

	acc, err := a.provider.AccountsService.GetAccount(ctx, identifier, password)
	if err != nil {
		return nil, err
	}

	if acc == nil {
		return nil, errors.New("no account found for identifier or email")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(password)); err != nil {
		return nil, errors.New("password does not match")
	}

	identity, err := a.provider.IdentityService.FetchIdentityByAccountID(ctx, acc.ID)
	if err != nil {
		return nil, err
	}

	tokenVal := &gwt.GWT[*gwt.Resources]{
		Header: gwt.Header{
			Issuer:    a.issuer,
			Recipient: []byte(identifier),
			Expires:   time.Now().Add(gwt.TokenExpireTime),
		},
		Body: identity.Resources,
	}

	tok, err := a.mc.Encode(tokenVal)
	if err != nil {
		return nil, err
	}

	identity.Account.Signature = tok.Signature
	tokenVal.Token = tok.Token

	err = a.provider.IdentityService.Updates(ctx, identity)
	if err != nil {
		return nil, err
	}

	return tokenVal, nil
}
func (a *Authentity) LoginToken(tkn string) error {
	// Validate Token
	tokenVal, err := a.mc.Decode(tkn)
	if err != nil {
		return err
	}

	return tokenVal.ValidateGWT(&a.spice)
}
func (a *Authentity) LogoutToken(pCtx context.Context, tkn string) error {
	tokenVal, err := a.mc.Decode(tkn)
	if err != nil {
		return err
	}

	if err = tokenVal.ValidateGWT(&a.spice); err != nil {
		return err
	}

	account, e := a.provider.AccountsService.FindAccountByUsername(pCtx, string(tokenVal.Header.Recipient))
	if e != nil {
		return errors.New("account not found")
	}

	account.Signature = ""

	defer a.l.TRACE("account just logged out", account)
	return a.provider.AccountsService.Updates(pCtx, account)
}

func (a *Authentity) RefreshToken(tkn string) (gwt.Token, error) {
	t, err := a.mc.Decode(tkn)
	if err != nil {
		return gwt.Token{}, err
	}

	if err = t.ValidateGWTHeader(func(header *gwt.Header) error {
		if !bytes.Equal(header.Issuer, a.issuer) {
			return errors.New("invalid token issuer")
		}

		return a.provider.AccountsService.AccountHasUsername(context.Background(), string(header.Recipient))
	}, &a.spice); err != nil {
		return gwt.Token{}, err
	}

	return a.mc.Encode(&gwt.GWT[*gwt.Resources]{
		Header: gwt.Header{
			Issuer:    a.issuer,
			Recipient: t.Header.Recipient,
			Expires:   time.Now().Add(gwt.TokenExpireTime),
		},
		Body: t.Body,
	})
}

func RegisterHandler(service *Authentity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest
		var err error

		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, service.l.ERROR(err.Error()), http.StatusBadRequest)
			return
		}

		err = service.RegisterIdentity(r.Context(), &req.Profile, &req.Account)
		if err != nil {
			http.Error(w, service.l.ERROR(err.Error()), http.StatusInternalServerError)
			return
		}

		service.l.INFO(req.Account.Email + " has been registered")
	}
}
func LoginHandler(service *Authentity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, service.l.ERROR(err.Error()), http.StatusBadRequest)
			return
		}

		var identifier string
		if req.Username != "" {
			identifier = req.Username
		} else if req.Email != "" {
			identifier = req.Email
		} else {
			service.l.ERROR("invalid login.html request")
			http.Error(w, "invalid login.html request", http.StatusBadRequest)
			return
		}

		if req.Password == "" {
			service.l.ERROR("password is required")
			http.Error(w, "password is required", http.StatusBadRequest)
			return
		}

		tokenValue, err := service.LoginManual(r.Context(), identifier, req.Password)
		if err != nil {
			service.l.ERROR(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    CookieTokenName,
			Value:   tokenValue.Token,
			Expires: tokenValue.Header.Expires,
			MaxAge:  0,
		})

		service.l.INFO(req.Email + "has logged in")
	}
}
func LogoutHandler(service *Authentity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(CookieTokenName)
		if err != nil {
			service.l.ERROR(err.Error())
			http.Error(w, "token not found", http.StatusUnauthorized)
			return
		}

		err = service.LogoutToken(r.Context(), tokenCookie.Value)
		if err != nil {
			service.l.ERROR(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    CookieTokenName,
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
		})

		service.l.INFO("token: " + tokenCookie.Value + " has logged out")
	}
}
