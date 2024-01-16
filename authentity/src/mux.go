package src

import (
	"context"
	"github.com/vaiktorg/grimoire/authentity/src/handlers"
	"github.com/vaiktorg/grimoire/gwt"
	"net/http"
	"time"
)

func (a *Authentity) registerMux() {

	a.mux.HandleFunc("/register", RegisterHandler(a))
	a.mux.HandleFunc("/login", LoginHandler(a))
	a.mux.HandleFunc("/logout", LogoutHandler(a))

	a.mux.Handle("/account", a.AuthenticationAndAuthorizationMiddleware(
		gwt.DataManagementUserData,
		gwt.DefaultRoles[gwt.Owner],
		handlers.AccountHandler(&a.provider.AccountsService)),
	)

	a.mux.Handle("/accounts", a.AuthenticationAndAuthorizationMiddleware(
		gwt.DataManagementUserData,
		gwt.DefaultRoles[gwt.Owner],
		handlers.AccountsHandler(&a.provider.AccountsService)),
	)

	a.mux.Handle("/profile", a.AuthenticationAndAuthorizationMiddleware(
		gwt.DataManagementUserData,
		gwt.DefaultRoles[gwt.Owner],
		handlers.ProfileHandler(&a.provider.ProfileService)),
	)

	a.mux.Handle("/profiles", a.AuthenticationAndAuthorizationMiddleware(
		gwt.DataManagementUserData,
		gwt.DefaultRoles[gwt.Owner],
		handlers.ProfilesHandler(&a.provider.ProfileService)),
	)

	a.mux.Handle("/identity", a.AuthenticationAndAuthorizationMiddleware(
		gwt.DataManagementUserData,
		gwt.DefaultRoles[gwt.Owner],
		handlers.IdentityHandler(&a.provider.IdentityService)),
	)

	a.mux.Handle("/identities", a.AuthenticationAndAuthorizationMiddleware(
		gwt.DataManagementUserData,
		gwt.DefaultRoles[gwt.Owner],
		handlers.IdentitiesHandler(&a.provider.IdentityService)),
	)

	a.l.TRACE("mux paths registered")
}

func TokenMiddleware(service *Authentity, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie(CookieTokenName)
		if err != nil {
			service.l.ERROR("token not found")
			http.Error(w, "token not found", http.StatusUnauthorized)
			return
		}

		if tokenCookie.Value == "" {
			service.l.ERROR("token value not found")
			http.Error(w, "token value not found", http.StatusUnauthorized)
			return
		}

		if time.Now().UTC().After(tokenCookie.Expires) {
			service.l.ERROR("token expired")
			http.Error(w, "token expired", http.StatusUnauthorized)
			return
		}

		if err = service.LoginToken(tokenCookie.Value); err != nil {
			service.l.ERROR(err.Error(), "Redirecting to /login.html")
			http.Redirect(w, r, "/auth/login.html", http.StatusTemporaryRedirect)
			return
		}

		ctx := context.WithValue(r.Context(), "token", tokenCookie.Value)

		next.ServeHTTP(w, r.WithContext(ctx))
		service.l.INFO("token: " + tokenCookie.Value + " has authed token")
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

			service.l.INFO("Resource Access", tok.Body.Resources)

			// User has the required access, proceed with the request
			next.ServeHTTP(w, r)
		}
	}
}

func (a *Authentity) AuthenticationAndAuthorizationMiddleware(resType gwt.ResourceType, role gwt.Role, next http.Handler) http.Handler {
	return TokenMiddleware(a, ResourceAccessMiddleware(a, resType, role)(next))
}
