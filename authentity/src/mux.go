package src

import (
	"context"
	"github.com/vaiktorg/grimoire/authentity/src/handlers"
	"github.com/vaiktorg/grimoire/gwt"
	"net/http"
	"time"
)

func (a *Authentity) registerMux() {

	a.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "tmpl/home.html")
	})
	a.mux.HandleFunc("/register", RegisterHandler(a))
	a.mux.HandleFunc("/login", LoginHandler(a))
	a.mux.HandleFunc("/logout", LogoutHandler(a))

	a.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("res/"))))

	a.mux.Handle("/account", a.AuthMiddleware(
		gwt.DataManagement,
		gwt.DefaultRoles[gwt.Owner],
		handlers.AccountHandler(&a.Provider.AccountsService)),
	)

	a.mux.Handle("/accounts", a.AuthMiddleware(
		gwt.DataManagement,
		gwt.DefaultRoles[gwt.Owner],
		handlers.AccountsHandler(&a.Provider.AccountsService)),
	)

	a.mux.Handle("/profile", a.AuthMiddleware(
		gwt.DataManagement,
		gwt.DefaultRoles[gwt.Owner],
		handlers.ProfileHandler(&a.Provider.ProfileService)),
	)

	a.mux.Handle("/profiles", a.AuthMiddleware(
		gwt.DataManagement,
		gwt.DefaultRoles[gwt.Owner],
		handlers.ProfilesHandler(&a.Provider.ProfileService)),
	)

	a.mux.Handle("/identity", a.AuthMiddleware(
		gwt.DataManagement,
		gwt.DefaultRoles[gwt.Owner],
		handlers.IdentityHandler(&a.Provider.IdentityService)),
	)

	a.mux.Handle("/identities", a.AuthMiddleware(
		gwt.DataManagement,
		gwt.DefaultRoles[gwt.Owner],
		handlers.IdentitiesHandler(&a.Provider.IdentityService)),
	)

	a.Logger.TRACE("mux paths registered")
}

func TokenMiddleware(service *Authentity, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(CookieTokenName)
		if err != nil {
			service.Logger.ERROR("token not found")
			http.Error(w, "token not found", http.StatusUnauthorized)
			return
		}

		if tokenCookie.Value == "" {
			service.Logger.ERROR("token value not found")
			http.Error(w, "token value not found", http.StatusUnauthorized)
			return
		}

		if time.Now().UTC().After(tokenCookie.Expires) {
			service.Logger.ERROR("token expired")
			http.Error(w, "token expired", http.StatusUnauthorized)
			return
		}

		if err = service.LoginToken(tokenCookie.Value); err != nil {
			service.Logger.ERROR(err.Error(), "Redirecting to /login.html")
			http.Redirect(w, r, "/auth/login.html", http.StatusTemporaryRedirect)
			return
		}

		ctx := context.WithValue(r.Context(), "token", tokenCookie.Value)

		next.ServeHTTP(w, r.WithContext(ctx))
		service.Logger.INFO("token: " + tokenCookie.Value + " has authed token")
	}
}
func ResourceAccessMiddleware(service *Authentity, resType gwt.ResourceType, roles ...gwt.Role) func(next http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Step 1: Extract user information from the request (e.g., from a tokStr or session)
			tokStr, ok := r.Context().Value("tokStr").(string)
			if !ok {
				http.Error(w, "could not fetch tokStr for authorization", http.StatusInternalServerError)
				return
			}

			// Step 2: Fetch user's resources and permissions
			tok, err := service.mc.Decode(tokStr)
			if err != nil {
				// Handle error: user not found, resources not found, etc.
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			// Step 3: Check if the user has access to the required resource with the required permission
			if !tok.Body.HasAccess(resType, roles...) {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			service.Logger.INFO("Resource Access", tok.Body.Resources)

			// User has the required access, proceed with the request
			next.ServeHTTP(w, r)
		}
	}
}

func (a *Authentity) AuthMiddleware(resType gwt.ResourceType, role gwt.Role, next http.Handler) http.Handler {
	return TokenMiddleware(a, ResourceAccessMiddleware(a, resType, role)(next))
}
