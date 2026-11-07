package agentwork

import (
	"context"
	"errors"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"gorm.io/gorm"
)

// ErrForeignFacileID is returned when the facile_id already names an entry that belongs
// to neither the owner nor one of the owner's agents.
var ErrForeignFacileID = errors.New("facile_id belongs to another user")

// Session is one finished span of agent work, booked to the owner's agent user.
type Session struct {
	OwnerID     int64
	Agent       schemas.AgentIdentity
	FacileID    string
	Facts       attribution.Facts
	StartedAt   time.Time
	StoppedAt   time.Time
	PausedMs    int64
	Cowork      bool
	Description string
}

// Result is what Record booked: the entry, the agent user it belongs to and its target.
type Result struct {
	Entry  schemas.TimeEntry
	Agent  schemas.User
	Target Target
}

// Record attributes the session through resolver, then ensures the owner's agent user and
// upserts the entry by facile_id in one transaction. The resolver runs first and outside
// the transaction, since its AI step may wait on the network. Replaying the same session
// updates the entry in place, except where the owner already reassigned it in Review.
func Record(ctx context.Context, db *gorm.DB, resolver attribution.Resolver, session Session) (*Result, error) {
	owner, err := attribution.LoadOwner(db, session.OwnerID)
	if err != nil {
		return nil, err
	}
	facts := session.Facts
	facts.At = session.StartedAt
	resolved, err := resolver.Resolve(ctx, db, owner, []attribution.Facts{facts})
	if err != nil {
		return nil, err
	}
	var result Result
	err = db.Transaction(func(tx *gorm.DB) error {
		agent, err := schemas.EnsureAgentUser(tx, session.OwnerID, session.Agent)
		if err != nil {
			return err
		}
		target, err := bookTarget(tx, resolved[0], facts.ProjectHint)
		if err != nil {
			return err
		}
		entry, err := upsertEntry(tx, agent.ID, *target, session)
		if err != nil {
			return err
		}
		result = Result{Entry: *entry, Agent: *agent, Target: *target}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func upsertEntry(tx *gorm.DB, agentID int64, target Target, session Session) (*schemas.TimeEntry, error) {
	stopped := session.StoppedAt.UTC()
	var entry schemas.TimeEntry
	err := tx.Where("facile_id = ?", session.FacileID).First(&entry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		facileID := session.FacileID
		entry = schemas.TimeEntry{FacileID: &facileID, LegacyDescription: session.Description}
	} else if err != nil {
		return nil, err
	} else if err := claimable(tx, entry.UserID, session.OwnerID, agentID); err != nil {
		return nil, err
	}
	if !reviewed(entry) {
		entry.ProjectID = target.Project.ID
		entry.TaskID = target.Task.ID
		entry.SpaceID = target.Project.SpaceID
		entry.Attribution = target.Attribution
	}
	entry.UserID = agentID
	entry.StartedAt = session.StartedAt.UTC()
	entry.StoppedAt = &stopped
	entry.PausedDurationMs = session.PausedMs
	entry.Cowork = session.Cowork
	if session.Description != "" {
		entry.LegacyDescription = session.Description
	}
	if err := tx.Save(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func reviewed(entry schemas.TimeEntry) bool {
	return entry.AttributionSource != nil && *entry.AttributionSource == attribution.SourceManual
}

func claimable(tx *gorm.DB, holderID, ownerID, agentID int64) error {
	if holderID == agentID || holderID == ownerID {
		return nil
	}
	var count int64
	err := tx.Model(&schemas.User{}).
		Where("id = ? AND kind = ? AND owner_user_id = ?", holderID, schemas.UserKindAgent, ownerID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrForeignFacileID
	}
	return nil
}
