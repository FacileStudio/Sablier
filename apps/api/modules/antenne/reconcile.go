package antenne

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	enveloppe "github.com/FacileStudio/enveloppe/go"
	"gorm.io/gorm"
)

// Reconcile assigns missing facile_ids, then queues `created` for every synced project,
// every task on one, and every time entry on such a task, in that order. Receivers treat
// a known facile_id as an update, so running it again is harmless.
func (s *Service) Reconcile(ctx context.Context) (*SyncResult, error) {
	if !s.shouldEmit() {
		return nil, errors.Failed("antenne is not enabled")
	}
	orm := s.orm.WithContext(ctx)
	if err := schemas.BackfillFacileIDs(orm); err != nil {
		return nil, errors.Internal("failed to assign facile ids", err)
	}
	var projects []schemas.Project
	if err := orm.Order("id").Find(&projects).Error; err != nil {
		return nil, errors.Internal("failed to list projects", err)
	}
	result := &SyncResult{}
	for i := range projects {
		if !s.emitProject(enveloppe.ActionCreated, &projects[i]) {
			continue
		}
		result.ProjectsSynced++
		tasks, entries := s.resendProject(orm, &projects[i])
		result.TasksSynced += tasks
		result.TimeEntriesSynced += entries
	}
	return result, nil
}

// ReconcileAtBoot runs Reconcile once and logs the outcome. It does nothing while the
// pool is not configured.
func (s *Service) ReconcileAtBoot(ctx context.Context) {
	if !s.shouldEmit() {
		return
	}
	result, err := s.Reconcile(ctx)
	if err != nil {
		s.logger.Warn("antenne: boot reconcile failed", slog.Any("error", err))
		return
	}
	s.logger.Info("antenne: boot reconcile queued",
		slog.Int("projects", result.ProjectsSynced),
		slog.Int("tasks", result.TasksSynced),
		slog.Int("time_entries", result.TimeEntriesSynced))
}

// resendProject queues `created` for each task of a synced project and each of their
// time entries, and reports how many of each it queued.
func (s *Service) resendProject(orm *gorm.DB, project *schemas.Project) (int, int) {
	var tasks []schemas.Task
	orm.Where("project_id = ?", project.ID).Order("id").Find(&tasks)
	sent, entries := 0, 0
	for i := range tasks {
		if s.emitTask(enveloppe.ActionCreated, &tasks[i], project) {
			sent++
			entries += s.resendEntries(orm, tasks[i].ID)
		}
	}
	return sent, entries
}

func (s *Service) resendEntries(orm *gorm.DB, taskID int64) int {
	var entries []schemas.TimeEntry
	orm.Where("task_id = ?", taskID).Order("id").Find(&entries)
	for i := range entries {
		s.EmitTimeEntryEvent(enveloppe.ActionCreated, &entries[i])
	}
	return len(entries)
}

// resendAdopted re-queues what was sent under an id the other side ignored: the time
// entries of an adopted task, or the tasks and entries of an adopted project.
func (s *Service) resendAdopted(projectID, taskID int64) {
	if taskID != 0 {
		s.resendEntries(s.orm, taskID)
	}
	var project schemas.Project
	if projectID == 0 || s.orm.Where("id = ?", projectID).First(&project).Error != nil {
		return
	}
	s.resendProject(s.orm, &project)
}
