package attribution

import (
	"context"
	"fmt"
	"strconv"

	"github.com/FacileStudio/Sablier/apps/api/internal/decide"
	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"gorm.io/gorm"
)

const (
	maxCandidates  = 30
	personalLabel  = "Personal"
	instructions   = "Pick the task this tracked work belongs to, from its title, project hint and file."
	questionPrefix = "work_"
)

type aiBatch struct {
	facts   []Facts
	spaces  []*string
	open    []int
	results []Result
}

type candidate struct {
	schemas.Task
	ProjectName string
	SpaceName   *string
}

func (c candidate) label() string {
	space := personalLabel
	if c.SpaceName != nil {
		space = *c.SpaceName
	}
	return fmt.Sprintf("%s › %s › %s", space, c.ProjectName, c.Name)
}

func candidates(db *gorm.DB, ownerID int64, space *string) ([]candidate, error) {
	query := visibleTasks(db, ownerID).
		Select("tasks.*, projects.name AS project_name, spaces.name AS space_name").
		Joins("LEFT JOIN spaces ON spaces.id = CAST(projects.space_id AS TEXT)").
		Where("tasks.status <> ?", schemas.StatusDone)
	if space != nil {
		query = query.Where("CAST(projects.space_id AS TEXT) = ?", *space)
	}
	var out []candidate
	err := query.Order("tasks.updated_at desc").Order("tasks.id desc").Limit(maxCandidates).Find(&out).Error
	return out, err
}

func labelled(tasks []candidate) ([]string, map[string]schemas.Task) {
	labels := make([]string, 0, len(tasks))
	byLabel := make(map[string]schemas.Task, len(tasks))
	for _, task := range tasks {
		label := task.label()
		if _, taken := byLabel[label]; taken {
			label += " #" + strconv.FormatInt(task.ID, 10)
		}
		labels = append(labels, label)
		byLabel[label] = task.Task
	}
	return labels, byLabel
}

func entryFor(db *gorm.DB, ownerID int64, batch aiBatch, i int) (entry, bool) {
	item := batch.facts[i]
	if item.Title == "" && item.ProjectHint == "" && item.Entity == "" {
		return entry{}, false
	}
	tasks, err := candidates(db, ownerID, batch.spaces[i])
	if err != nil || len(tasks) == 0 {
		return entry{}, false
	}
	labels, byLabel := labelled(tasks)
	return newEntry(i, item, labels, byLabel)
}

func buildPrompts(db *gorm.DB, ownerID int64, batch aiBatch) []prompt {
	entries := make([]entry, 0, len(batch.open))
	for _, i := range batch.open {
		if e, ok := entryFor(db, ownerID, batch, i); ok {
			entries = append(entries, e)
		}
	}
	return split(entries)
}

func (r Resolver) ask(ctx context.Context, db *gorm.DB, owner Owner, batch aiBatch) {
	if r.Decider == nil || len(batch.open) == 0 {
		return
	}
	for _, p := range buildPrompts(db, owner.ID, batch) {
		answers, err := r.Decider.Decide(ctx, owner.Email, p.state, p.questions)
		if err != nil {
			continue
		}
		for name, answer := range answers {
			if result := pick(db, p.asked[name], answer); result != nil {
				batch.results[p.asked[name].index] = *result
			}
		}
	}
}

func pick(db *gorm.DB, asked question, answer decide.Answer) *Result {
	task, ok := asked.byLabel[answer.Choice]
	if !ok {
		return nil
	}
	result, err := withProject(db, SourceAI, min(max(answer.Confidence, 0), 1), task)
	if err != nil {
		return nil
	}
	return result
}
