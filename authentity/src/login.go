package src

import (
	"encoding/json"
	"github.com/vaiktorg/grimoire/authentity/src/models"
	"net/http"
)

func LoginHandler(service *Authentity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getLogin(service, w, r)
		case http.MethodPost:
			postLogin(service, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func getLogin(service *Authentity, w http.ResponseWriter, r *http.Request) {
	service.Logger.INFO("login.page.gohtml request received ", r.RemoteAddr+" "+r.Method+" "+r.URL.Path)
	http.ServeFile(w, r, "tmpl/login.page.gohtml")
}

func postLogin(service *Authentity, w http.ResponseWriter, r *http.Request) {

	var req models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, service.Logger.ERROR(err.Error()), http.StatusBadRequest)
		return
	}

	var identifier string
	if req.Username != "" {
		identifier = req.Username
	} else if req.Email != "" {
		identifier = req.Email
	} else {
		service.Logger.ERROR("invalid login.page.gohtml request")
		http.Error(w, "invalid login.page.gohtml request", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		service.Logger.ERROR("password is required")
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	tokenValue, err := service.LoginManual(r.Context(), identifier, req.Password)
	if err != nil {
		service.Logger.ERROR(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    CookieTokenName,
		Value:   tokenValue.Token,
		Expires: tokenValue.Header.Expires,
		MaxAge:  0,
	})

	service.Logger.INFO(req.Email + "has logged in")
}
