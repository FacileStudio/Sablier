package decide

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/FacileStudio/porte/exchange"
	"github.com/FacileStudio/tronc/bulle"
)

const (
	// BackendBulle names the suite backend, Bulle's POST /api/ai/decide.
	BackendBulle = "bulle"
	// BackendOff names the absence of a backend: the AI step is skipped.
	BackendOff = "off"

	timeout = 5 * time.Second
)

var errNoUser = errors.New("decide: no user on the context")

type userKey struct{}

// TokenFunc returns the Bulle token exchanged for the user with this id.
type TokenFunc func(ctx context.Context, userID int64) (string, error)

// Config holds what the backend is chosen from: Bulle's URL and the way to get a user's
// exchanged Bulle token.
type Config struct {
	BulleURL string
	Token    TokenFunc
}

// ConfigFromEnv reads BULLE_URL. The caller sets Token.
func ConfigFromEnv() Config {
	return Config{BulleURL: strings.TrimSpace(os.Getenv("BULLE_URL"))}
}

// New returns Bulle's decider when both BulleURL and Token are set, else a nil Decider,
// which turns the AI step off. It also returns the backend's name to log, and an error
// when BulleURL is not an http or https origin.
func New(cfg Config) (Decider, string, error) {
	if cfg.BulleURL == "" || cfg.Token == nil {
		return nil, BackendOff, nil
	}
	client, err := bulle.New(bulle.Config{
		URL: cfg.BulleURL,
		TokenSource: func(ctx context.Context) (string, error) {
			userID, ok := ctx.Value(userKey{}).(int64)
			if !ok {
				return "", errNoUser
			}
			return cfg.Token(ctx, userID)
		},
		HTTPClient: &http.Client{Timeout: timeout},
	})
	if err != nil {
		return nil, BackendOff, err
	}
	return service{client: client}, BackendBulle, nil
}

type service struct {
	client *bulle.Client
}

// Decide asks Bulle as the user with this id, with the token exchanged for them, so
// Bulle checks that user's own plan. A user who must sign in again gets no answer.
func (s service) Decide(ctx context.Context, userID int64, state any, questions map[string]Question) (map[string]Answer, error) {
	ctx = context.WithValue(ctx, userKey{}, userID)
	decision, err := s.client.DecideAI(ctx, bulle.DecideRequest{State: state, Questions: questions})
	if err != nil {
		if errors.Is(err, exchange.ErrReconnect) {
			slog.DebugContext(ctx, "decide: user has no Registre grant, AI step skipped", slog.Int64("user_id", userID))
		}
		return nil, err
	}
	return decision.Answers, nil
}
