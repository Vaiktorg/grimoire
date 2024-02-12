package handlers

import (
	"encoding/json"
	"github.com/vaiktorg/grimoire/authentity/src/services"
	"github.com/vaiktorg/grimoire/uid"
	"net/http"
)

// ====================================================================================================
// Handlers

func ProfilesHandler(service *services.ProfileService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pass the request context onto the database layer.
		bks, err := service.AllProfiles(r.Context())
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
func ProfileHandler(service *services.ProfileService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pass the request context onto the database layer.
		id := r.URL.Query().Get("id")

		if id == "" {
			http.Error(w, "id not provided", http.StatusInternalServerError)
			return
		}

		bks, err := service.FetchProfile(r.Context(), uid.UID(id))
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
