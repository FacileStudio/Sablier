package decide

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"sync"
	"time"
)

const (
	answerTTL  = 10 * time.Minute
	failureTTL = time.Minute
	maxCached  = 512
)

type cached struct {
	answers map[string]Answer
	err     error
	expires time.Time
}

type cache struct {
	next    Decider
	now     func() time.Time
	mu      sync.Mutex
	entries map[[32]byte]cached
}

// Cached wraps next so the same user asking the same questions about the same state
// is answered from memory: answers for ten minutes, failures for one, so a slow or
// missing model does not add its timeout to every read of the drafts.
func Cached(next Decider) Decider {
	return &cache{next: next, now: time.Now, entries: map[[32]byte]cached{}}
}

func (c *cache) Decide(ctx context.Context, userEmail string, state any, questions map[string]Question) (map[string]Answer, error) {
	key, err := keyOf(userEmail, state, questions)
	if err != nil {
		return c.next.Decide(ctx, userEmail, state, questions)
	}
	now := c.now()
	c.mu.Lock()
	hit, ok := c.entries[key]
	c.mu.Unlock()
	if ok && now.Before(hit.expires) {
		return hit.answers, hit.err
	}
	answers, err := c.next.Decide(ctx, userEmail, state, questions)
	ttl := answerTTL
	if err != nil {
		ttl = failureTTL
	}
	c.store(key, cached{answers: answers, err: err, expires: now.Add(ttl)}, now)
	return answers, err
}

func (c *cache) store(key [32]byte, entry cached, now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= maxCached {
		for k, v := range c.entries {
			if !now.Before(v.expires) {
				delete(c.entries, k)
			}
		}
	}
	if len(c.entries) >= maxCached {
		c.entries = map[[32]byte]cached{}
	}
	c.entries[key] = entry
}

func keyOf(userEmail string, state any, questions map[string]Question) ([32]byte, error) {
	body, err := json.Marshal([]any{userEmail, state, questions})
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(body), nil
}
