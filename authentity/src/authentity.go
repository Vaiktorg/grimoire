package src

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"github.com/vaiktorg/grimoire/authentity/src/services"
	"github.com/vaiktorg/grimoire/gwt"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/uid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type Authentity struct {
	Issuer   string
	Spice    gwt.Spice
	Provider *DataProvider
	encoder  *gwt.Encoder[*gwt.GWT[AuthBody]]
	decoder  *gwt.Decoder[*gwt.GWT[AuthBody]]
	logger   log.ISimLogger
}

func NewAuthentity(issuerName string, logger log.ISimLogger, dialer gorm.Dialector) *Authentity {
	db, err := gorm.Open(dialer, &gorm.Config{})
	if err != nil {
		panic(err)
	}

	spice := gwt.Spice{
		Salt:   uid.NewUID(8).Bytes(),
		Pepper: uid.NewUID(8).Bytes(),
	}

	encoder := gwt.NewEncoder[*gwt.GWT[AuthBody]](spice)
	decoder := gwt.NewDecoder[*gwt.GWT[AuthBody]](spice)
	auth := &Authentity{
		Provider: NewDataProvider(db),
		Issuer:   issuerName,
		encoder:  &encoder,
		decoder:  &decoder,
		Spice:    spice,
		logger:   logger,
	}

	if err = auth.Migrate(); err != nil && !errors.Is(err, AlreadyExistError) {
		panic(err)
	}

	return auth
}

func (a *Authentity) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()

	mux.HandleFunc("/auth/register", RegisterHandler(a))
	mux.HandleFunc("/auth/login", LoginHandler(a))
	mux.HandleFunc("/auth/logout", LogoutHandler(a))

	mux.HandleFunc("/auth/account", LoginTokenMiddleware(a, services.AccountHandler(&a.Provider.AccountsService)))
	mux.HandleFunc("/auth/accounts", LoginTokenMiddleware(a, services.AccountsHandler(&a.Provider.AccountsService)))

	mux.HandleFunc("/auth/profile", LoginTokenMiddleware(a, services.ProfileHandler(&a.Provider.ProfileService)))
	mux.HandleFunc("/auth/profiles", LoginTokenMiddleware(a, services.ProfilesHandler(&a.Provider.ProfileService)))

	mux.HandleFunc("/auth/identity", LoginTokenMiddleware(a, services.IdentityHandler(&a.Provider.IdentityService)))
	mux.HandleFunc("/auth/identities", LoginTokenMiddleware(a, services.IdentitiesHandler(&a.Provider.IdentityService)))

	mux.ServeHTTP(w, r)
}

func RegisterHandler(service *Authentity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		var err error

		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			service.logger.ERROR(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = service.RegisterIdentity(r.Context(), &req.Profile, &req.Account)
		if err != nil {
			service.logger.ERROR(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		service.logger.INFO(req.Account.Email + " has been registered")
	}
}
func LoginHandler(service *Authentity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		var err error

		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			service.logger.ERROR(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var identifier string
		if req.Username != "" {
			identifier = req.Username
		} else if req.Email != "" {
			identifier = req.Email
		} else {
			service.logger.ERROR(err.Error())
			http.Error(w, "invalid login request", http.StatusBadRequest)
			return
		}

		if req.Password == "" {
			service.logger.ERROR(err.Error())
			http.Error(w, "password is required", http.StatusBadRequest)
			return
		}

		tokenValue, err := service.LoginManual(r.Context(), identifier, req.Password)
		if err != nil {
			service.logger.ERROR(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    CookieTokenName,
			Value:   tokenValue.Token.Token,
			Expires: tokenValue.Header.Expires,
			MaxAge:  0,
		})

		service.logger.INFO(req.Email + "has logged in")
	}
}
func LogoutHandler(service *Authentity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(CookieTokenName)
		if err != nil {
			service.logger.ERROR(err.Error())
			http.Error(w, "token not found", http.StatusUnauthorized)
			return
		}

		err = service.LogoutToken(r.Context(), tokenCookie.Value)
		if err != nil {
			service.logger.ERROR(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Del("Set-Cookie")

		service.logger.INFO("token: " + tokenCookie.Value + " has logged out")
	}
}
func LoginTokenMiddleware(service *Authentity, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(CookieTokenName)
		if err != nil {
			service.logger.ERROR(err.Error())
			http.Error(w, "token not found", http.StatusUnauthorized)
			return
		}

		if time.Since(tokenCookie.Expires) >= gwt.TokenExpireTime {
			service.logger.ERROR(err.Error())
			http.Error(w, "token expired", http.StatusUnauthorized)
			return
		}

		if tokenCookie.Value == "" {
			service.logger.ERROR(err.Error())
			http.Error(w, "token value not found", http.StatusUnauthorized)
			return
		}

		if err = service.LoginToken(r.Context(), tokenCookie.Value); err != nil {
			service.logger.ERROR(err.Error())
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)

		service.logger.INFO("token: " + tokenCookie.Value + " has authed token")
	}
}

func (a *Authentity) RegisterIdentity(pCtx context.Context, prof *entities.Profile, acc *entities.Account) error {
	ctx, cancel := context.WithTimeout(pCtx, time.Minute)
	defer cancel()

	if _, err := a.Provider.IdentityService.FetchIdentityByEmail(ctx, acc.Email); err == nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("email already being used for email: " + acc.Email)
	}
	if _, err := a.Provider.IdentityService.FetchIdentityByUsername(ctx, acc.Username); err == nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("username already being used for user: " + acc.Username)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(acc.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("could not create hashed password")
	}

	tok := &gwt.GWT[AuthBody]{
		Header: gwt.Header{
			Issuer:    a.Issuer,
			Recipient: acc.Username,
			Expires:   time.Now().Add(gwt.TokenExpireTime),
		},
	}

	t, err := a.encoder.Encode(tok)
	if err != nil {
		return err
	}

	identity := &entities.Identity{
		Profile: prof,
		Account: &entities.Account{
			Username: acc.Username,
			Email:    acc.Email,
			Password: string(hashedPassword),
		},
		Signature: t.Signature,
	}

	return a.Provider.IdentityService.Repo.Persist(ctx, identity)
}
func (a *Authentity) LoginManual(pCtx context.Context, identifier, password string) (*gwt.GWT[AuthBody], error) {
	ctx, cancel := context.WithTimeout(pCtx, time.Minute)
	defer cancel()

	acc, err := a.Provider.AccountsService.GetAccount(ctx, identifier, password)
	if err != nil {
		return nil, err
	}

	if acc == nil {
		return nil, errors.New("no account found for identifier or email")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(password)); err != nil {
		return nil, errors.New("password does not match")
	}

	identity, err := a.Provider.IdentityService.FetchIdentityByAccountID(ctx, acc.ID)
	if err != nil {
		return nil, err
	}

	val := gwt.GWT[AuthBody]{
		Header: gwt.Header{
			Issuer:    a.Issuer,
			Recipient: identifier,
			Expires:   time.Now().Add(gwt.TokenExpireTime),
		},
	}

	tok, err := a.encoder.Encode(&val)
	if err != nil {
		return nil, err
	}

	identity.Signature = nil
	identity.Signature = tok.Signature

	val.Token.Token = tok.Token
	val.Token.Signature = tok.Signature

	err = a.Provider.IdentityService.Persist(ctx, identity)
	if err != nil {
		return nil, err
	}

	return &val, nil
}
func (a *Authentity) LoginToken(pCtx context.Context, tkn string) error {
	// Validate Token
	tokenVal, err := a.decoder.Decode(tkn)
	if err != nil {
		return err
	}

	if err = tokenVal.ValidateGWT(nil); err != nil {
		return err
	}

	clearSig := func(ctx context.Context, identity *entities.Identity) error {
		identity.Signature = nil
		return a.Provider.IdentityService.Persist(ctx, identity)
	}

	// Validate Session ----------------------------------------------------------------------------------------------------
	ctx, cancel := context.WithTimeout(pCtx, time.Minute)
	defer cancel()

	identity, err := a.Provider.IdentityService.FetchIdentityByUsername(ctx, tokenVal.Header.Recipient)
	if err != nil {
		return err
	}

	if time.Since(tokenVal.Header.Expires) >= gwt.TokenExpireTime {
		err = errors.New("token expired")
		return errors.Join(err, clearSig(ctx, identity))
	}

	if !bytes.Equal(tokenVal.Token.Signature, identity.Signature) {
		err = errors.New(gwt.ErrorInvalidTokenSignature)
		return errors.Join(err, clearSig(ctx, identity))
	}

	return nil
}
func (a *Authentity) LogoutToken(pCtx context.Context, tkn string) error {
	tokenVal, err := a.decoder.Decode(tkn)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(pCtx, time.Minute)
	defer cancel()

	if err = tokenVal.ValidateGWT(nil); err != nil {
		return err
	}

	account, e := a.Provider.IdentityService.FetchIdentityByUsername(ctx, tokenVal.Header.Recipient)
	if e != nil {
		return errors.New("account not found")
	}

	identity, e := a.Provider.IdentityService.FetchIdentity(ctx, account.ID)
	if e != nil {
		return errors.New("account not found")
	}

	identity.Signature = nil

	return a.Provider.IdentityService.Persist(pCtx, identity)
}
