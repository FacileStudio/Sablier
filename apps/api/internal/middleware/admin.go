package middleware

import (
	"net/http"
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// Admins is the SABLIER_ADMINS allow-list: the emails that may read and change
// Sablier's app-wide settings (the webhook and the Antenne bus). Sablier has no
// admin role of its own and any Registre account can sign in, so an empty list
// closes those routes to everyone.
type Admins []string

// Allows reports whether email is on the list, ignoring case and surrounding space.
func (admins Admins) Allows(email string) bool {
	want := strings.ToLower(strings.TrimSpace(email))
	if want == "" {
		return false
	}
	for _, admin := range admins {
		if strings.ToLower(strings.TrimSpace(admin)) == want {
			return true
		}
	}
	return false
}

// RequireAdmin is RequireAuth followed by a 403 for any account whose email is
// not on admins. The email is the account's own: the sign-in rewrites it from
// Registre's verified address, and PATCH /users/me cannot change it.
func RequireAdmin(auth Authenticator, admins Admins) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return RequireAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			identity, ok := authcontext.IdentityFromContext(request.Context())
			if !ok || !admins.Allows(identity.Email) {
				httpjson.WriteError(w, errors.Forbidden("administrator access required"))
				return
			}
			next.ServeHTTP(w, request)
		}))
	}
}
