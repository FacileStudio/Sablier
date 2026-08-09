package auth

import documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "auth",
	Description: "Authentication, through the suite's porte kit. A session is a random 32 byte token stored SHA-256 hashed, sent as Authorization: Bearer <token> or, after a single sign-on login, as an HttpOnly __Host-session cookie. A cookie-authenticated request that is not GET, HEAD or OPTIONS must also carry an X-Facile-CSRF header with any non-empty value.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/auth/logout",
			Summary:      "End the current session",
			Description:  "Served by porte. Revokes the one session this request authenticated with, by id, so it cannot end somebody else's, and clears the session cookie.",
			Auth:         "bearer",
			ResponseBody: `{"logged_out":true}`,
		},
		{
			Method:      "GET",
			Path:        "/auth/oidc",
			Summary:     "Start the single sign-on flow",
			Description: "Registered only when OIDC_ISSUER is set. Redirects to the identity provider with PKCE, a nonce and a state, all three carried in one short-lived HttpOnly cookie — none of which the hand-written client it replaces had. Add ?flow=cli (and ?port=N when listening on loopback) to end on a one-time code instead of a cookie.",
			Auth:        "public",
		},
		{
			Method:      "GET",
			Path:        "/auth/oidc/callback",
			Summary:     "Finish the single sign-on flow",
			Description: "The provider's redirect target. Verifies the state in constant time, the ID token and its nonce, upserts the account, sets the session cookie and redirects to OIDC_SUCCESS_URL. No token reaches the URL any more: it used to ride back in the fragment.",
			Auth:        "public",
		},
		{
			Method:       "POST",
			Path:         "/auth/oidc/exchange",
			Summary:      "Trade a CLI login code for a session token",
			Description:  "The other half of ?flow=cli. The code is single use, hashed at rest and valid for 60 seconds.",
			Auth:         "public",
			RequestBody:  `{"code":string}`,
			ResponseBody: `{"user_id":string,"token":string}`,
		},
		{
			Method:       "POST",
			Path:         "/auth/backchannel-logout",
			Summary:      "Revoke every session for a user, on the provider's instruction",
			Description:  "OpenID Connect Back-Channel Logout 1.0, called by the identity provider rather than by a client. Note the deployed Authentik is 2025.6.3, which has no field to configure this, so nothing calls it yet.",
			Auth:         "public",
			ResponseBody: `{"logged_out":true}`,
		},
		{
			Method:      "POST",
			Path:        "/auth/sync-profile",
			Summary:     "Refresh the profile from the identity provider",
			Description: "Rate limited to one call per user per five minutes; synced is false when the window had not elapsed.",
			Auth:        "bearer",
		},
		{
			Method:      "GET",
			Path:        "/auth/config",
			Summary:     "Describe the auth methods on offer",
			Description: "Returns sso_only and oidc_enabled, so the client knows whether to show the password form.",
			Auth:        "public",
		},
		{
			Method:       "POST",
			Path:         "/auth/register",
			Summary:      "Register a new user",
			Description:  "Creates a user account and returns an auth token.",
			Auth:         "public",
			RequestBody:  "RegisterRequest",
			ResponseBody: "AuthResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or invalid registration input."},
				{Status: 409, Code: "already_exists", Description: "A user with the same email already exists."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "POST",
			Path:         "/auth/login",
			Summary:      "Authenticate a user",
			Description:  "Authenticates credentials and returns an auth token.",
			Auth:         "public",
			RequestBody:  "LoginRequest",
			ResponseBody: "AuthResponse",
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or invalid login input."},
				{Status: 401, Code: "unauthenticated", Description: "Email or password is invalid."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
	},
}
