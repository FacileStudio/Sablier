package antenne

import (
	"fmt"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"

	enveloppe "github.com/FacileStudio/enveloppe/go"
)

func (s *Service) emitting(channel string) bool {
	return s.shouldEmit() && s.IsPoolEventEnabled(channel)
}

func (s *Service) enqueue(object enveloppe.ObjectType, action enveloppe.Action, facileID string, payload any) {
	channel := fmt.Sprintf("%s.%s", object, action)
	now := time.Now().UTC()
	s.enqueueOutbox(channel, enveloppe.Event[any]{
		Version:        enveloppe.EventVersion,
		App:            enveloppe.AppSablier,
		Object:         object,
		Action:         action,
		FacileID:       facileID,
		Payload:        payload,
		Timestamp:      now.Format(time.RFC3339),
		IdempotencyKey: fmt.Sprintf("%s:%s:%d", channel, facileID, now.UnixNano()),
	})
}

func (s *Service) ensureFacileID(model any, id int64, facileID **string) {
	if *facileID != nil {
		return
	}
	fid := GenerateFacileID()
	*facileID = &fid
	s.orm.Model(model).Where("id = ?", id).UpdateColumn("facile_id", fid)
}

// markLinked records that a live change went out, so a later twin with another id from
// the other side is not merged into this row.
func (s *Service) markLinked(model any, id int64) {
	s.orm.Model(model).Where("id = ? AND synced_at IS NULL", id).UpdateColumn("synced_at", time.Now().UTC())
}

// EmitProjectEvent queues a live project change for the pool. A project in a space
// Registre does not know stays local.
func (s *Service) EmitProjectEvent(action enveloppe.Action, project *schemas.Project) {
	if s.emitProject(action, project) {
		s.markLinked(&schemas.Project{}, project.ID)
	}
}

func (s *Service) emitProject(action enveloppe.Action, project *schemas.Project) bool {
	if !s.emitting("project." + string(action)) {
		return false
	}
	space, ok := s.spaceOf(project)
	if !ok {
		return false
	}
	p := *project
	s.ensureFacileID(&schemas.Project{}, p.ID, &p.FacileID)
	project.FacileID = p.FacileID
	s.enqueue(enveloppe.ObjectProject, action, *p.FacileID, projectPayload{
		FacileID:    *p.FacileID,
		Name:        p.Name,
		Description: p.Description,
		Icon:        normalizeIcon(p.Icon),
		Space:       space,
		UpdatedAt:   wireTime(p.UpdatedAt),
	})
	return true
}

// EmitTaskEvent queues a live task change for the pool, after its project when the
// project had never been given a facile_id.
func (s *Service) EmitTaskEvent(action enveloppe.Action, task *schemas.Task, project *schemas.Project) {
	if s.emitTask(action, task, project) {
		s.markLinked(&schemas.Task{}, task.ID)
	}
}

func (s *Service) emitTask(action enveloppe.Action, task *schemas.Task, project *schemas.Project) bool {
	if project == nil || !s.emitting("task."+string(action)) {
		return false
	}
	if _, ok := s.spaceOf(project); !ok {
		return false
	}
	p := *project
	if p.FacileID == nil {
		s.EmitProjectEvent(enveloppe.ActionCreated, &p)
		s.ensureFacileID(&schemas.Project{}, p.ID, &p.FacileID)
	}
	t := *task
	s.ensureFacileID(&schemas.Task{}, t.ID, &t.FacileID)
	s.enqueue(enveloppe.ObjectTask, action, *t.FacileID, s.taskPayload(&t, *p.FacileID))
	return true
}

func (s *Service) taskPayload(t *schemas.Task, projectFacileID string) taskPayload {
	out := taskPayload{
		FacileID:        *t.FacileID,
		ProjectFacileID: projectFacileID,
		Name:            t.Name,
		Status:          schemas.NormalizeStatus(t.Status),
		Description:     t.Description,
		DueDate:         t.DueDate,
		Priority:        t.Priority,
		UpdatedAt:       wireTime(t.UpdatedAt),
	}
	if t.SyncedName != "" {
		out.Name = t.SyncedName
	}
	if out.Priority == "" {
		out.Priority = "none"
	}
	if t.ActorID != nil {
		out.ActorEmail, out.ActorSubject = s.identityOf(*t.ActorID)
	}
	return out
}

// EmitTimeEntryEvent queues a time entry change. Only entries on a task of a synced
// project leave Sablier; Opus mirrors them read-only.
func (s *Service) EmitTimeEntryEvent(action enveloppe.Action, entry *schemas.TimeEntry) {
	if !s.emitting("time_entry." + string(action)) {
		return
	}
	payload, ok := s.timeEntryPayload(entry)
	if !ok {
		return
	}
	s.enqueue(enveloppe.ObjectTimeEntry, action, payload.FacileID, payload)
}

func (s *Service) timeEntryPayload(entry *schemas.TimeEntry) (timeEntryPayload, bool) {
	var task schemas.Task
	var project schemas.Project
	if entry.TaskID == 0 || s.orm.Where("id = ?", entry.TaskID).First(&task).Error != nil {
		return timeEntryPayload{}, false
	}
	if s.orm.Where("id = ?", task.ProjectID).First(&project).Error != nil {
		return timeEntryPayload{}, false
	}
	if _, ok := s.spaceOf(&project); !ok || project.FacileID == nil {
		return timeEntryPayload{}, false
	}
	e := *entry
	s.ensureFacileID(&schemas.TimeEntry{}, e.ID, &e.FacileID)
	s.ensureFacileID(&schemas.Task{}, task.ID, &task.FacileID)
	out := timeEntryPayload{
		FacileID:         *e.FacileID,
		TaskFacileID:     *task.FacileID,
		ProjectFacileID:  *project.FacileID,
		StartedAt:        e.StartedAt.UTC().Format(time.RFC3339),
		PausedDurationMs: e.PausedDurationMs,
		Description:      e.LegacyDescription,
		UpdatedAt:        wireTime(e.UpdatedAt),
	}
	out.UserEmail, out.UserSubject = s.identityOf(e.UserID)
	if e.StoppedAt != nil {
		stopped := e.StoppedAt.UTC().Format(time.RFC3339)
		out.StoppedAt = &stopped
	}
	return out, true
}
