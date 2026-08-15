package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/FacileStudio/porte"
	troncenv "github.com/FacileStudio/tronc/env"
	"github.com/joho/godotenv"
)

// OIDCConfig holds the OIDC settings read from the environment when
// OIDC_ISSUER is set.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	SuccessURL   string
}

// Porte returns the shared auth kit's configuration. The environment variable
// names are the suite convention and do not change; this only reshapes what
// Load already read.
//
// SSOOnly rides along because porte serves /auth/config and the frontend
// decides on it whether to draw a password form at all.
func (c Config) Porte() porte.Config {
	if c.OIDC == nil {
		return porte.Config{SSOOnly: c.SSOOnly}
	}
	return porte.Config{
		Issuer:       c.OIDC.Issuer,
		ClientID:     c.OIDC.ClientID,
		ClientSecret: c.OIDC.ClientSecret,
		RedirectURL:  c.OIDC.RedirectURL,
		SuccessURL:   c.OIDC.SuccessURL,
		SSOOnly:      c.SSOOnly,
		ClaimsScope:  c.OIDCClaimsScope,
	}
}

// Config is the fully-loaded application configuration produced by Load.
type Config struct {
	troncenv.Core
	StorageDir      string
	OIDC            *OIDCConfig
	OIDCClaimsScope string
	SSOOnly         bool
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string

	// JournalBrowserURL and JournalBrowserKey configure the browser's error
	// reporting. Both empty leaves the client reporting nothing.
	//
	// The URL is deliberately not JOURNAL_URL. That one is the server SDK's,
	// and it is documented as http://journal-api:4010 — a Docker-internal
	// address a browser cannot resolve. Handing it to the client would
	// produce a page that reports diligently into nowhere.
	//
	// The key is a Journal *public* ingest key and is public by
	// construction: what bounds its abuse is the origin allowlist and the
	// daily quota carried by the key itself, not its secrecy.
	JournalBrowserURL string
	JournalBrowserKey string
}

// Load reads the .env file (if present) and environment variables into a
// Config, validating PORT, LOG_LEVEL, and, when OIDC_ISSUER is set, the
// required OIDC_CLIENT_ID/OIDC_CLIENT_SECRET/OIDC_REDIRECT_URL.
//
// JOURNAL_BROWSER_URL is rejected unless it ends in /api: Journal's
// dashboard answers any unmatched path with 200 and an HTML document, so a
// base URL without /api would silently swallow every browser error report
// behind a 2xx that reads as success. Load refuses that configuration at
// boot rather than let it surface later as reports that never arrive.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("failed to parse .env file: %w", err)
	}

	core, err := troncenv.LoadCore()
	if err != nil {
		return Config{}, err
	}
	if core.Port < 1 || core.Port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a valid TCP port")
	}
	if err := validateLogLevel(core.LogLevel); err != nil {
		return Config{}, err
	}

	ssoOnly, err := troncenv.Bool("SSO_ONLY", false)
	if err != nil {
		return Config{}, err
	}

	env := Config{
		Core:            core,
		StorageDir:      troncenv.String("STORAGE_DIR", "./data"),
		SSOOnly:         ssoOnly,
		VAPIDPublicKey:  troncenv.String("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey: troncenv.String("VAPID_PRIVATE_KEY", ""),
		VAPIDSubject:    troncenv.String("VAPID_SUBJECT", "mailto:admin@example.com"),
		OIDCClaimsScope: troncenv.String("OIDC_CLAIMS_SCOPE", ""),

		JournalBrowserURL: troncenv.String("JOURNAL_BROWSER_URL", ""),
		JournalBrowserKey: troncenv.String("JOURNAL_BROWSER_KEY", ""),
	}

	if env.JournalBrowserURL != "" && !strings.HasSuffix(strings.TrimRight(env.JournalBrowserURL, "/"), "/api") {
		return Config{}, fmt.Errorf("JOURNAL_BROWSER_URL must end in /api, got %q", env.JournalBrowserURL)
	}

	if issuer := troncenv.String("OIDC_ISSUER", ""); issuer != "" {
		clientID := troncenv.String("OIDC_CLIENT_ID", "")
		clientSecret := troncenv.String("OIDC_CLIENT_SECRET", "")
		redirectURL := troncenv.String("OIDC_REDIRECT_URL", "")
		if clientID == "" || clientSecret == "" || redirectURL == "" {
			return Config{}, fmt.Errorf("OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL are required when OIDC_ISSUER is set")
		}
		successURL := troncenv.String("OIDC_SUCCESS_URL", "")
		if successURL == "" && len(env.CORSAllowedOrigins) > 0 {
			successURL = env.CORSAllowedOrigins[0]
		}
		env.OIDC = &OIDCConfig{
			Issuer:       issuer,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			SuccessURL:   successURL,
		}
	}

	return env, nil
}

func validateLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error")
	}
}
