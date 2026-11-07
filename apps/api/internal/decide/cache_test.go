package decide

import (
	"context"
	"errors"
	"testing"
	"time"
)

type counting struct {
	calls int
	err   error
}

func (c *counting) Decide(context.Context, string, any, map[string]Question) (map[string]Answer, error) {
	c.calls++
	if c.err != nil {
		return nil, c.err
	}
	return map[string]Answer{"q0": {Choice: "a", Confidence: 0.9}}, nil
}

func TestCachedAnswersRepeatsFromMemory(t *testing.T) {
	inner := &counting{}
	clock := time.Unix(0, 0)
	decider := Cached(inner).(*cache)
	decider.now = func() time.Time { return clock }
	questions := map[string]Question{"q0": {Type: "choice", Options: []string{"a", "b"}}}
	ask := func(state string) {
		if _, err := decider.Decide(context.Background(), "a@example.test", state, questions); err != nil {
			t.Fatal(err)
		}
	}
	ask("x")
	ask("x")
	if inner.calls != 1 {
		t.Fatalf("a repeated question reached the model %d times", inner.calls)
	}
	ask("y")
	clock = clock.Add(answerTTL)
	ask("x")
	if inner.calls != 3 {
		t.Fatalf("a new state and an expired answer must ask again, got %d calls", inner.calls)
	}
	inner.err = errors.New("down")
	if _, err := decider.Decide(context.Background(), "a@example.test", "z", questions); err == nil {
		t.Fatal("want the failure")
	}
	if _, err := decider.Decide(context.Background(), "a@example.test", "z", questions); err == nil || inner.calls != 4 {
		t.Fatalf("a recent failure must be served from memory, got %d calls", inner.calls)
	}
}
