// Package decide asks Bulle's POST /api/ai/decide to pick among options, as Sablier's
// service acting for a user. Bulle is the only way to the decision model: without
// BULLE_URL and BULLE_SERVICE_KEY there is no AI step. Any failure, a timeout, an older
// Bulle's 404 or a 402 for a user without AI included, is an error the caller reads as
// "no answer".
package decide

import (
	"context"

	"github.com/FacileStudio/tronc/bulle"
)

// Question is one typed ask. A "choice" question carries its Options and a Criteria
// object keyed by option that says what belongs to each: the decision model answers 422
// to a choice without it.
type Question = bulle.Question

// Answer is one question's result: the picked option, its calibrated confidence and the
// probability of every option.
type Answer = bulle.Answer

// Decider answers questions about a state on behalf of the user named by userEmail.
type Decider interface {
	Decide(ctx context.Context, userEmail string, state any, questions map[string]Question) (map[string]Answer, error)
}
