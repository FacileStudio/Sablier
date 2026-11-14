package attribution

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/decide"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

type recordingDecider struct {
	states    []any
	questions []map[string]decide.Question
}

func (r *recordingDecider) Decide(_ context.Context, _ string, state any, questions map[string]decide.Question) (map[string]decide.Answer, error) {
	r.states = append(r.states, state)
	r.questions = append(r.questions, questions)
	return map[string]decide.Answer{}, nil
}

func TestAIRequestsFitBulleDecideContract(t *testing.T) {
	w := seedWorld(t)
	long := strings.Repeat("x", 180)
	for i := range 40 {
		create(t, w.db, &schemas.Task{ProjectID: w.named.ID, Name: long + strconv.Itoa(i)})
	}
	facts := make([]Facts, 20)
	for i := range facts {
		facts[i] = Facts{Title: "work " + strconv.Itoa(i)}
	}
	facts = append(facts, Facts{Title: "single", RegistreSpaceID: "reg-team"})
	recorder := &recordingDecider{}
	if _, err := (Resolver{Decider: recorder, Access: &access{ai: true}}).Resolve(context.Background(), w.db, w.owner, facts); err != nil {
		t.Fatal(err)
	}
	if len(recorder.questions) < 3 {
		t.Fatalf("21 open items must be split over several calls, got %d", len(recorder.questions))
	}
	asked := 0
	for call, questions := range recorder.questions {
		asked += checkCall(t, recorder.states[call], questions)
	}
	if asked != len(facts) {
		t.Fatalf("every open item must be asked once, got %d of %d", asked, len(facts))
	}
}

func checkCall(t *testing.T, state any, questions map[string]decide.Question) int {
	t.Helper()
	encodedState, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(map[string]any{"state": state, "questions": questions})
	if err != nil {
		t.Fatal(err)
	}
	if len(questions) == 0 || len(questions) > 8 || len(encodedState) > 32<<10 || len(body) > 128<<10 {
		t.Fatalf("a call breaks Bulle's limits: questions=%d state=%dB body=%dB", len(questions), len(encodedState), len(body))
	}
	for name, q := range questions {
		checkChoice(t, name, q)
	}
	return len(questions)
}

func checkChoice(t *testing.T, name string, q decide.Question) {
	t.Helper()
	wire, err := json.Marshal(q)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Type     string            `json:"type"`
		Criteria map[string]string `json:"criteria"`
		Options  []string          `json:"options"`
	}
	if err := json.Unmarshal(wire, &decoded); err != nil {
		t.Fatalf("%s: criteria must encode as an object keyed by option: %v %s", name, err, wire)
	}
	if decoded.Type != "choice" || len(decoded.Options) < 2 || len(decoded.Options) > 64 || len(decoded.Criteria) != len(decoded.Options) {
		t.Fatalf("%s: want a choice with 2 to 64 options and one criterion each: %s", name, wire)
	}
	for _, option := range decoded.Options {
		if decoded.Criteria[option] == "" {
			t.Fatalf("%s: option %q has no criterion", name, option)
		}
	}
}
