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

type question struct {
	index   int
	byLabel map[string]schemas.Task
}

type prompt struct {
	state     map[string]any
	questions map[string]decide.Question
	asked     map[string]question
}

func buildPrompt(db *gorm.DB, ownerID int64, batch aiBatch) prompt {
	p := prompt{state: map[string]any{}, questions: map[string]decide.Question{}, asked: map[string]question{}}
	for _, i := range batch.open {
		item := batch.facts[i]
		if item.Title == "" && item.ProjectHint == "" && item.Entity == "" {
			continue
		}
		tasks, err := candidates(db, ownerID, batch.spaces[i])
		if err != nil || len(tasks) == 0 {
			continue
		}
		labels, byLabel := labelled(tasks)
		name := questionPrefix + strconv.Itoa(i)
		p.state[name] = map[string]any{"title": item.Title, "project_hint": item.ProjectHint, "entity": item.Entity, "candidates": labels}
		p.questions[name] = decide.Question{Type: "choice", Instructions: instructions, Options: labels}
		p.asked[name] = question{index: i, byLabel: byLabel}
	}
	return p
}

func (r Resolver) ask(ctx context.Context, db *gorm.DB, owner Owner, batch aiBatch) {
	if r.Decider == nil || len(batch.open) == 0 {
		return
	}
	p := buildPrompt(db, owner.ID, batch)
	if len(p.questions) == 0 {
		return
	}
	answers, err := r.Decider.Decide(ctx, owner.Email, p.state, p.questions)
	if err != nil {
		return
	}
	for name, answer := range answers {
		if result := pick(db, p.asked[name], answer); result != nil {
			batch.results[p.asked[name].index] = *result
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
