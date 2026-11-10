package middleware

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/porte/kit"
)

// Authenticator is the auth service: porte's session middleware, plus the
// lookup that turns the user id porte resolved into the identity the rest of
// Sablier reads. It stays one parameter so every module router keeps calling
// middleware.RequireAuth(authService).
type Authenticator interface {
	kit.Authenticator
	kit.Resolver
}

// RequireAuth is porte/kit's middleware, which authenticates the cookie or
// bearer credential and hydrates the identity, followed by the one step the
// kit leaves to Sablier: its handlers read the user id as a decimal string
// from authcontext, so the identity is copied there.
func RequireAuth(auth Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return kit.RequireAuth(auth, auth)(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			identity, _ := kit.FromContext(request.Context())
			ctx := authcontext.WithIdentity(request.Context(), authcontext.Identity{
				UserID: strconv.FormatInt(identity.UserID, 10),
				Email:  identity.Email,
			})
			next.ServeHTTP(w, request.WithContext(ctx))
		}))
	}
}
