package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type Payload struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

func Fire(url, secretHeader, secretValue string, payload Payload) {
	go func() {
		body, err := json.Marshal(payload)
		if err != nil {
			slog.Error("webhook: failed to marshal payload", slog.Any("error", err))
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			slog.Error("webhook: failed to create request", slog.Any("error", err))
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if secretHeader != "" && secretValue != "" {
			req.Header.Set(secretHeader, secretValue)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("webhook: request failed", slog.Any("error", err))
			return
		}
		defer resp.Body.Close()
		slog.Info("webhook: fired", slog.String("event", payload.Event), slog.Int("status", resp.StatusCode))
	}()
}
