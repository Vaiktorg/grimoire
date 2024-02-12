package handlers

import (
	"encoding/json"
	"github.com/vaiktorg/grimoire/authentity/src/services"
	"net/http"
)

// ====================================================================================================
// AccountsHandlers

func AccountsHandler(service *services.AccountService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bks, err := service.AllAccounts(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = json.NewEncoder(w).Encode(bks)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	}
}
func AccountHandler(service *services.AccountService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info := &struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			ID       string `json:"id"`
		}{}

		err := json.NewDecoder(r.Body).Decode(info)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}

		account, err := service.GetAccount(r.Context(), info.Username, info.Email)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = json.NewEncoder(w).Encode(account)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	}
}
