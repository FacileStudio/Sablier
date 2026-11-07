package agents

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/agentwork"
	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

const (
	maxNameLength        = 80
	maxFacileIDLength    = 200
	maxDescriptionLength = 500
	maxAvatarURLLength   = 2048
	maxFactLength        = 200
	maxEntityLength      = 1024
	maxOpusTaskIDs       = 50
	maxSessionLength     = 24 * time.Hour
	maxClockSkew         = 5 * time.Minute
)

var agentKeyPattern = regexp.MustCompile(`^[a-z0-9-]{1,40}$`)

// toSession validates a request at the trust boundary and turns it into the
// session agentwork books for ownerID.
func toSession(ownerID int64, req *RecordSessionRequest, now time.Time) (agentwork.Session, error) {
	identity, err := toIdentity(req)
	if err != nil {
		return agentwork.Session{}, err
	}
	facileID := strings.TrimSpace(req.FacileID)
	if facileID == "" || len(facileID) > maxFacileIDLength {
		return agentwork.Session{}, errors.Invalid("facile_id is required and at most 200 characters")
	}
	if len(req.Description) > maxDescriptionLength {
		return agentwork.Session{}, errors.Invalid("description is at most 500 characters")
	}
	facts, err := toFacts(req)
	if err != nil {
		return agentwork.Session{}, err
	}
	pausedMs, err := pausedMillis(req, now)
	if err != nil {
		return agentwork.Session{}, err
	}
	return agentwork.Session{
		OwnerID:     ownerID,
		Agent:       identity,
		FacileID:    facileID,
		Facts:       facts,
		StartedAt:   req.StartedAt.UTC(),
		StoppedAt:   req.StoppedAt.UTC(),
		PausedMs:    pausedMs,
		Cowork:      req.Cowork,
		Description: strings.TrimSpace(req.Description),
	}, nil
}

func toFacts(req *RecordSessionRequest) (attribution.Facts, error) {
	if req.ProjectID < 0 || req.TaskID < 0 {
		return attribution.Facts{}, errors.Invalid("project_id and task_id must be positive")
	}
	if len(req.OpusTaskIDs) > maxOpusTaskIDs {
		return attribution.Facts{}, errors.Invalid("opus_task_ids holds at most 50 ids")
	}
	short := append([]string{req.Project, req.OpusUpdatedTaskID, req.OpusProjectID, req.RegistreSpaceID,
		req.OpusWorkspace, req.Title}, req.OpusTaskIDs...)
	for _, value := range short {
		if len(value) > maxFactLength {
			return attribution.Facts{}, errors.Invalid("project, title and the opus and registre ids are at most 200 characters")
		}
	}
	if len(req.Entity) > maxEntityLength {
		return attribution.Facts{}, errors.Invalid("entity is at most 1024 characters")
	}
	return attribution.Facts{
		ProjectID:         req.ProjectID,
		TaskID:            req.TaskID,
		OpusTaskIDs:       req.OpusTaskIDs,
		OpusUpdatedTaskID: strings.TrimSpace(req.OpusUpdatedTaskID),
		OpusProjectID:     strings.TrimSpace(req.OpusProjectID),
		RegistreSpaceID:   strings.TrimSpace(req.RegistreSpaceID),
		OpusWorkspace:     strings.TrimSpace(req.OpusWorkspace),
		Title:             strings.TrimSpace(req.Title),
		ProjectHint:       strings.TrimSpace(req.Project),
		Entity:            strings.TrimSpace(req.Entity),
	}, nil
}

func toIdentity(req *RecordSessionRequest) (schemas.AgentIdentity, error) {
	if !agentKeyPattern.MatchString(req.AgentKey) {
		return schemas.AgentIdentity{}, errors.Invalid("agent_key must match [a-z0-9-]{1,40}")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > maxNameLength {
		return schemas.AgentIdentity{}, errors.Invalid("name is required and at most 80 characters")
	}
	avatar := strings.TrimSpace(req.AvatarURL)
	if avatar == "" {
		avatar = schemas.BlobAvatarURL(name)
	} else if !isHTTPSURL(avatar) {
		return schemas.AgentIdentity{}, errors.Invalid("avatar_url must be an https URL")
	}
	return schemas.AgentIdentity{Key: req.AgentKey, Name: name, AvatarURL: avatar}, nil
}

func isHTTPSURL(raw string) bool {
	if len(raw) > maxAvatarURLLength {
		return false
	}
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" && parsed.User == nil
}

func pausedMillis(req *RecordSessionRequest, now time.Time) (int64, error) {
	if req.StartedAt == nil || req.StoppedAt == nil {
		return 0, errors.Invalid("started_at and stopped_at are required")
	}
	span := req.StoppedAt.Sub(*req.StartedAt)
	if span < 0 || span > maxSessionLength {
		return 0, errors.Invalid("stopped_at must follow started_at by at most 24 hours")
	}
	if req.StoppedAt.After(now.Add(maxClockSkew)) {
		return 0, errors.Invalid("stopped_at is in the future")
	}
	if req.ActiveSeconds == nil {
		return 0, nil
	}
	if *req.ActiveSeconds < 0 {
		return 0, errors.Invalid("active_seconds must not be negative")
	}
	active := time.Duration(*req.ActiveSeconds) * time.Second
	if active >= span {
		return 0, nil
	}
	return (span - active).Milliseconds(), nil
}
