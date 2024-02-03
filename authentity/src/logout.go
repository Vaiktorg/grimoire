package src

import (
	"net/http"
	"time"
)

func LogoutHandler(service *Authentity) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(CookieTokenName)
		if err != nil {
			service.Logger.ERROR(err.Error())
			http.Error(w, "token not found", http.StatusUnauthorized)
			return
		}

		err = service.LogoutToken(r.Context(), tokenCookie.Value)
		if err != nil {
			service.Logger.ERROR(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    CookieTokenName,
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
		})

		service.Logger.INFO("token: " + tokenCookie.Value + " has logged out")
	}
}
