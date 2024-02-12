package handlers

import (
	"encoding/json"
	"github.com/vaiktorg/grimoire/authentity/src/services"
	"net/http"
)

func IdentityHandler(service *services.IdentityService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pass the request context onto the database layer.
		id := r.URL.Query().Get("id")

		bks, err := service.FetchIdentity(r.Context(), id)
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
func IdentitiesHandler(service *services.IdentityService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pass the request context onto the database layer.
		bks, err := service.FetchIdentities(r.Context())
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
