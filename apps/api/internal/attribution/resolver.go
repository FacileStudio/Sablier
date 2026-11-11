// Package attribution decides which project and task a piece of tracked work belongs to.
// It tries, in order: explicit ids, an Opus task, the project hint, the owner's running
// entry, an AI choice among the space's tasks, then gives up with no project. Each step
// sets a confidence, and a result below the owner's threshold needs review. It only reads:
// the caller creates whatever task or catch-all project it books on.
package attribution

import (
	"context"
	"errors"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/decide"
	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"gorm.io/gorm"
)

const (
	// SourceExplicit is a visible task_id or project_id the caller sent.
	SourceExplicit = "explicit"
	// SourceOpus is an Opus task or project id that maps to a visible row.
	SourceOpus = "opus"
	// SourceName is a project hint that matches a visible project's name.
	SourceName = "name"
	// SourceRunning is the owner's own entry running when the work started.
	SourceRunning = "running"
	// SourceAI is the decision model's pick among the candidate tasks.
	SourceAI = "ai"
	// SourceFallback is no match at all.
	SourceFallback = "fallback"
	// SourceManual is an attribution the owner set or accepted in Review.
	SourceManual = "manual"

	// MinThreshold and MaxThreshold bound the user's review threshold.
	MinThreshold = 0.5
	MaxThreshold = 1.0
)

// ErrNotVisible is returned when an explicit project or task id is not one the owner can see.
var ErrNotVisible = errors.New("project or task is not visible to the owner")

// Owner is the user the work is attributed for, with the threshold that decides review.
type Owner struct {
	ID        int64
	Email     string
	Threshold float64
}

// Facts is everything a caller knows about the work. Every field is optional. At is when
// it started; a zero At skips the running-entry step.
type Facts struct {
	ProjectID         int64
	TaskID            int64
	OpusTaskIDs       []string
	OpusUpdatedTaskID string
	OpusProjectID     string
	RegistreSpaceID   string
	OpusWorkspace     string
	Title             string
	ProjectHint       string
	Entity            string
	At                time.Time
}

// Result is where the work goes. Project is nil for SourceFallback, and Task is nil when
// the step found a project but no task.
type Result struct {
	Project     *schemas.Project
	Task        *schemas.Task
	Source      string
	Confidence  float64
	NeedsReview bool
}

// Resolver runs the steps. A nil Decider skips the AI step.
type Resolver struct {
	Decider decide.Decider
}

// Resolve attributes every item of facts for owner, asking the decider once per batch of
// the items the rules left open, each batch within Bulle's decide limits.
func (r Resolver) Resolve(ctx context.Context, db *gorm.DB, owner Owner, facts []Facts) ([]Result, error) {
	results := make([]Result, len(facts))
	spaces := make([]*string, len(facts))
	var open []int
	for i, item := range facts {
		space, err := memberSpace(db, owner.ID, item)
		if err != nil {
			return nil, err
		}
		spaces[i] = space
		result, err := rules(db, owner.ID, item, space)
		if err != nil {
			return nil, err
		}
		if result == nil {
			open = append(open, i)
			continue
		}
		results[i] = *result
	}
	r.ask(ctx, db, owner, aiBatch{facts: facts, spaces: spaces, open: open, results: results})
	for i := range results {
		if results[i].Source == "" {
			results[i] = Result{Source: SourceFallback}
		}
		results[i].NeedsReview = results[i].Source == SourceFallback || results[i].Confidence < owner.Threshold
	}
	return results, nil
}

// LoadOwner reads the user's email and review threshold.
func LoadOwner(db *gorm.DB, userID int64) (Owner, error) {
	var user schemas.User
	if err := db.Select("id", "email", "review_threshold").First(&user, userID).Error; err != nil {
		return Owner{}, err
	}
	return Owner{ID: user.ID, Email: user.Email, Threshold: Threshold(user.ReviewThreshold)}, nil
}

// Threshold returns value when it lies within the allowed range, else the default.
func Threshold(value float64) float64 {
	if value < MinThreshold || value > MaxThreshold {
		return schemas.DefaultReviewThreshold
	}
	return value
}

func rules(db *gorm.DB, ownerID int64, item Facts, space *string) (*Result, error) {
	for _, run := range []step{explicit, opus, byName, running} {
		result, err := run(db, ownerID, item, space)
		if result != nil || err != nil {
			return result, err
		}
	}
	return nil, nil
}
