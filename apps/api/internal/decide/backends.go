package decide

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/FacileStudio/tronc/bulle"
)

const (
	// BackendBulle names the suite backend, Bulle's POST /api/ai/decide.
	BackendBulle = "bulle"
	// BackendOff names the absence of a backend: the AI step is skipped.
	BackendOff = "off"

	timeout = 5 * time.Second
)

// Config holds the environment the backend is chosen from.
type Config struct {
	BulleURL        string
	BulleServiceKey string
}

// ConfigFromEnv reads BULLE_URL and BULLE_SERVICE_KEY.
func ConfigFromEnv() Config {
	return Config{
		BulleURL:        strings.TrimSpace(os.Getenv("BULLE_URL")),
		BulleServiceKey: strings.TrimSpace(os.Getenv("BULLE_SERVICE_KEY")),
	}
}

// New returns Bulle's decider when both BulleURL and BulleServiceKey are set, else a nil
// Decider, which turns the AI step off. It also returns the backend's name to log, and an
// error when BulleURL is not an http or https origin.
func New(cfg Config) (Decider, string, error) {
	if cfg.BulleURL == "" || cfg.BulleServiceKey == "" {
		return nil, BackendOff, nil
	}
	key := cfg.BulleServiceKey
	client, err := bulle.New(bulle.Config{
		URL:         cfg.BulleURL,
		TokenSource: func(context.Context) (string, error) { return key, nil },
		HTTPClient:  &http.Client{Timeout: timeout},
	})
	if err != nil {
		return nil, BackendOff, err
	}
	return service{client: client}, BackendBulle, nil
}

type service struct {
	client *bulle.Client
}

// Decide asks Bulle as Sablier's service, acting for userEmail, whose plan Bulle checks.
func (s service) Decide(ctx context.Context, userEmail string, state any, questions map[string]Question) (map[string]Answer, error) {
	decision, err := s.client.DecideAI(ctx, bulle.DecideRequest{ActingFor: userEmail, State: state, Questions: questions})
	if err != nil {
		return nil, err
	}
	return decision.Answers, nil
}
