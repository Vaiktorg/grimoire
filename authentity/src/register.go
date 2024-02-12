package src

import (
	"encoding/json"
	"github.com/vaiktorg/grimoire/authentity/src/models"
	"net/http"
)

func RegisterHandler(service *Authentity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getRegister(service, w, r)
		case http.MethodPost:
			postRegister(service, w, r)
		default:
			http.Error(w, service.Logger.ERROR("method not allowed"), http.StatusMethodNotAllowed)
		}
	}
}

func getRegister(service *Authentity, w http.ResponseWriter, r *http.Request) {
	service.Logger.INFO("register page requested")
	http.ServeFile(w, r, "tmpl/register.html")
}

func postRegister(service *Authentity, w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	var err error

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, service.Logger.ERROR(err.Error()), http.StatusBadRequest)
		return
	}

	err = service.RegisterIdentity(r.Context(), &req.Profile, &req.Account)
	if err != nil {
		http.Error(w, service.Logger.ERROR(err.Error()), http.StatusInternalServerError)
		return
	}

	service.Logger.INFO(req.Account.Email + " has been registered")
}
