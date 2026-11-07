package decide

import (
	"context"
	"net/http"
	"os"
	"strings"
)

const (
	// BackendBulle names the suite backend, Bulle's POST /api/ai/decide.
	BackendBulle = "bulle"
	// BackendJev names the direct Jev backend of a standalone install.
	BackendJev = "jev"
	// BackendOff names the absence of a backend: the AI step is skipped.
	BackendOff = "off"

	jevDecisions       = "decisions"
	jevDefaultBase     = "https://api.typesafe.ai"
	jevDecisionsBase   = "https://openrouter.ai"
	jevDefaultModel    = "jev-latest"
	jevSystemOnePath   = "/v1/systemone"
	jevDecisionsPath   = "/api/alpha/decisions"
	bulleDecidePath    = "/api/ai/decide"
	bulleActingFor     = "X-Bulle-Acting-For"
	bearerPrefix       = "Bearer "
	authorizationField = "Authorization"
)

// Config holds the environment the backend is chosen from.
type Config struct {
	BulleURL        string
	BulleServiceKey string
	JevAPIKey       string
	JevEndpoint     string
	JevBaseURL      string
	JevModel        string
}

// ConfigFromEnv reads BULLE_URL, BULLE_SERVICE_KEY, JEV_API_KEY, JEV_ENDPOINT,
// JEV_BASE_URL and JEV_MODEL.
func ConfigFromEnv() Config {
	return Config{
		BulleURL:        strings.TrimSpace(os.Getenv("BULLE_URL")),
		BulleServiceKey: strings.TrimSpace(os.Getenv("BULLE_SERVICE_KEY")),
		JevAPIKey:       strings.TrimSpace(os.Getenv("JEV_API_KEY")),
		JevEndpoint:     strings.TrimSpace(os.Getenv("JEV_ENDPOINT")),
		JevBaseURL:      strings.TrimSpace(os.Getenv("JEV_BASE_URL")),
		JevModel:        strings.TrimSpace(os.Getenv("JEV_MODEL")),
	}
}

// New picks the backend once: Bulle when BulleURL is set, else direct Jev when JevAPIKey is
// set, else none, which returns a nil Decider. A Bulle install never falls back to Jev, so
// the data goes only where the install chose. It also returns the backend's name to log.
func New(cfg Config, client *http.Client) (Decider, string) {
	if client == nil {
		client = &http.Client{}
	}
	switch {
	case cfg.BulleURL != "":
		return bulle{caller: caller{http: client}, url: strings.TrimRight(cfg.BulleURL, "/") + bulleDecidePath,
			key: cfg.BulleServiceKey}, BackendBulle
	case cfg.JevAPIKey != "":
		return newJev(cfg, client), BackendJev
	default:
		return nil, BackendOff
	}
}

type bulle struct {
	caller
	url string
	key string
}

// Decide asks Bulle, acting for userEmail, with the service key as bearer.
func (b bulle) Decide(ctx context.Context, userEmail string, state any, questions map[string]Question) (map[string]Answer, error) {
	headers := map[string]string{bulleActingFor: userEmail}
	if b.key != "" {
		headers[authorizationField] = bearerPrefix + b.key
	}
	return b.post(ctx, b.url, headers, map[string]any{"state": state, "questions": questions})
}

type jev struct {
	caller
	url   string
	key   string
	model string
}

func newJev(cfg Config, client *http.Client) jev {
	path, base := jevSystemOnePath, jevDefaultBase
	if cfg.JevEndpoint == jevDecisions {
		path, base = jevDecisionsPath, jevDecisionsBase
	}
	if cfg.JevBaseURL != "" {
		base = strings.TrimRight(cfg.JevBaseURL, "/")
	}
	if strings.HasSuffix(base, "/api") && strings.HasPrefix(path, "/api/") {
		base = strings.TrimSuffix(base, "/api")
	}
	model := cfg.JevModel
	if model == "" {
		model = jevDefaultModel
	}
	return jev{caller: caller{http: client}, url: base + path, key: cfg.JevAPIKey, model: model}
}

// Decide calls Jev directly. userEmail is not sent: Jev knows no Sablier user.
func (j jev) Decide(ctx context.Context, _ string, state any, questions map[string]Question) (map[string]Answer, error) {
	headers := map[string]string{authorizationField: bearerPrefix + j.key}
	return j.post(ctx, j.url, headers, map[string]any{"state": state, "model": j.model, "questions": questions})
}
