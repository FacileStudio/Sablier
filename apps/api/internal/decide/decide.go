// Package decide asks a Jev-style decision model to pick among options. Bulle answers
// for the suite through POST /api/ai/decide; a Sablier deployed alone calls Jev itself.
// Both speak the same contract, {state, questions} in and {answers} out, so one set of
// types serves both. Any failure, a timeout, an older Bulle's 404 or a 402 for a user
// without AI included, is an error the caller reads as "no answer".
package decide

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	timeout = 5 * time.Second
	maxBody = 1 << 20
)

// Question is one typed ask. A "choice" question carries its Options and a Criteria
// object keyed by option that says what belongs to each: the decision model answers 422
// to a choice without it.
type Question struct {
	Type         string            `json:"type"`
	Instructions any               `json:"instructions"`
	Criteria     map[string]string `json:"criteria,omitempty"`
	Options      []string          `json:"options,omitempty"`
}

// Answer is one question's result: the picked option, its calibrated confidence and the
// probability of every option.
type Answer struct {
	Choice        string             `json:"choice"`
	Confidence    float64            `json:"confidence"`
	Probabilities map[string]float64 `json:"probabilities"`
}

// Decider answers questions about a state on behalf of the user named by userEmail.
type Decider interface {
	Decide(ctx context.Context, userEmail string, state any, questions map[string]Question) (map[string]Answer, error)
}

type response struct {
	Answers map[string]Answer `json:"answers"`
}

type caller struct {
	http *http.Client
}

func (c caller) post(ctx context.Context, url string, headers map[string]string, body any) (map[string]Answer, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(res.Body, maxBody))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("decide: %s answered %d", url, res.StatusCode)
	}
	var out response
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decide: malformed answer: %w", err)
	}
	return out.Answers, nil
}
