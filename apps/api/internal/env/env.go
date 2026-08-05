package env

import (
	"fmt"
	"os"
	"strings"

	troncenv "github.com/FacileStudio/tronc/env"
	"github.com/joho/godotenv"
)

type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	SuccessURL   string
}

type Config struct {
	troncenv.Core
	StorageDir      string
	OIDC            *OIDCConfig
	SSOOnly         bool
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string
}

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
