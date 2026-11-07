package heartbeats

import (
	"context"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/tronc/errors"
)

// ListDrafts returns userID's drafts in [from, to], each with the resolver's suggestion.
func (service *Service) ListDrafts(ctx context.Context, userID int64, from, to time.Time) ([]Draft, error) {
	return service.listDrafts(ctx, userID, from, to)
}

// DefaultWindow is the draft window GET /heartbeats/drafts uses when none is given.
func (service *Service) DefaultWindow() (time.Time, time.Time) {
	now := service.now()
	return now.Add(-defaultDraftWindow), now
}

func (service *Service) listDrafts(ctx context.Context, userID int64, from, to time.Time) ([]Draft, error) {
	spans, err := service.computeForUser(ctx, userID, from, to)
	if err != nil || len(spans) == 0 {
		return []Draft{}, err
	}
	db := service.orm.WithContext(ctx)
	owner, err := attribution.LoadOwner(db, userID)
	if err != nil {
		return nil, errors.Internal("failed to read the user", err)
	}
	facts := make([]attribution.Facts, len(spans))
	for i, s := range spans {
		facts[i] = s.facts()
	}
	results, err := service.resolver.Resolve(ctx, db, owner, facts)
	if err != nil {
		return nil, errors.Internal("failed to suggest projects for drafts", err)
	}
	drafts := make([]Draft, len(spans))
	for i, s := range spans {
		drafts[i] = toDraft(s)
		drafts[i].Suggestion = toSuggestion(results[i])
		if drafts[i].ProjectID == nil && drafts[i].Suggestion != nil {
			id := drafts[i].Suggestion.ProjectID
			drafts[i].ProjectID = &id
		}
	}
	return drafts, nil
}

func toSuggestion(result attribution.Result) *Suggestion {
	if result.Project == nil {
		return nil
	}
	suggestion := &Suggestion{
		ProjectID:   result.Project.ID,
		SpaceID:     result.Project.SpaceID,
		Source:      result.Source,
		Confidence:  result.Confidence,
		NeedsReview: result.NeedsReview,
	}
	if result.Task != nil {
		id := result.Task.ID
		suggestion.TaskID = &id
	}
	return suggestion
}

// facts collects what the span's heartbeats said: the explicit project when one sent it,
// the hint, and the latest space, Opus task, title and entity.
func (s span) facts() attribution.Facts {
	facts := attribution.Facts{ProjectHint: s.Hint}
	for _, b := range s.Beats {
		if b.Explicit && b.ProjectID != nil {
			facts.ProjectID = *b.ProjectID
		}
		facts.RegistreSpaceID = latest(facts.RegistreSpaceID, b.Facts.RegistreSpaceID)
		facts.Title = latest(facts.Title, b.Facts.Title)
		facts.Entity = latest(facts.Entity, b.Entity)
		if b.Facts.OpusTaskID != "" {
			facts.OpusTaskIDs = []string{b.Facts.OpusTaskID}
		}
	}
	return facts
}

func latest(current, next string) string {
	if next != "" {
		return next
	}
	return current
}
