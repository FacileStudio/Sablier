package attribution

import (
	"encoding/json"
	"strconv"

	"github.com/FacileStudio/Sablier/apps/api/internal/decide"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

const (
	maxQuestions  = 8
	maxStateBytes = 28 << 10
	maxBodyBytes  = 112 << 10
	noneLabel     = "None of these tasks"
	noneCriterion = "the work belongs to none of the listed tasks"
)

type question struct {
	index   int
	byLabel map[string]schemas.Task
}

type entry struct {
	name      string
	state     map[string]any
	question  decide.Question
	asked     question
	stateSize int
	bodySize  int
}

type prompt struct {
	state     map[string]any
	questions map[string]decide.Question
	asked     map[string]question
	stateSize int
	bodySize  int
}

func newEntry(i int, item Facts, labels []string, byLabel map[string]schemas.Task) (entry, bool) {
	state := map[string]any{"title": item.Title, "project_hint": item.ProjectHint, "entity": item.Entity, "candidates": labels}
	criteria := make(map[string]string, len(labels)+1)
	for _, label := range labels {
		criteria[label] = "the work is on the task " + label + " (space › project › task)"
	}
	criteria[noneLabel] = noneCriterion
	options := append(append(make([]string, 0, len(labels)+1), labels...), noneLabel)
	q := decide.Question{Type: "choice", Instructions: instructions, Criteria: criteria, Options: options}
	encodedState, stateErr := json.Marshal(state)
	encodedQuestion, questionErr := json.Marshal(q)
	if stateErr != nil || questionErr != nil {
		return entry{}, false
	}
	e := entry{name: questionPrefix + strconv.Itoa(i), state: state, question: q, asked: question{index: i, byLabel: byLabel},
		stateSize: len(encodedState), bodySize: len(encodedState) + len(encodedQuestion)}
	return e, e.stateSize <= maxStateBytes && e.bodySize <= maxBodyBytes
}

func newPrompt() prompt {
	return prompt{state: map[string]any{}, questions: map[string]decide.Question{}, asked: map[string]question{}}
}

func (p prompt) fits(e entry) bool {
	return len(p.questions) < maxQuestions && p.stateSize+e.stateSize <= maxStateBytes && p.bodySize+e.bodySize <= maxBodyBytes
}

func (p *prompt) add(e entry) {
	p.state[e.name], p.questions[e.name], p.asked[e.name] = e.state, e.question, e.asked
	p.stateSize += e.stateSize
	p.bodySize += e.bodySize
}

func split(entries []entry) []prompt {
	var prompts []prompt
	current := newPrompt()
	for _, e := range entries {
		if !current.fits(e) {
			prompts = append(prompts, current)
			current = newPrompt()
		}
		current.add(e)
	}
	if len(current.questions) > 0 {
		prompts = append(prompts, current)
	}
	return prompts
}
