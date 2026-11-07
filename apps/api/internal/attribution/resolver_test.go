package attribution

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/decide"
	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"gorm.io/gorm"
)

type stubDecider struct {
	choice     string
	confidence float64
	calls      int
}

func (s *stubDecider) Decide(_ context.Context, _ string, _ any, questions map[string]decide.Question) (map[string]decide.Answer, error) {
	s.calls++
	if s.choice == "" {
		return nil, stderrors.New("bulle answered 404")
	}
	answers := map[string]decide.Answer{}
	for name := range questions {
		answers[name] = decide.Answer{Choice: s.choice, Confidence: s.confidence}
	}
	return answers, nil
}

func create(t *testing.T, db *gorm.DB, values ...any) {
	t.Helper()
	for _, value := range values {
		if err := db.Create(value).Error; err != nil {
			t.Fatalf("seed %T: %v", value, err)
		}
	}
}

var start = time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)

type world struct {
	db                         *gorm.DB
	owner                      Owner
	team, named, inTeam, other schemas.Project
	opusTask, aiTask, runTask  schemas.Task
	hidden                     schemas.Project
}

func seedWorld(t *testing.T) world {
	t.Helper()
	db := testdb.Open(t)
	registre, teamFacile := "reg-team", "opus-task-1"
	space, foreign := schemas.Space{Name: "Facile", RegistreSpaceID: &registre}, schemas.Space{Name: "Elsewhere"}
	create(t, db, &schemas.User{ID: 1, Email: "sara@example.test"}, &space, &foreign)
	create(t, db, &schemas.SpaceMember{SpaceID: space.ID, UserID: 1})
	w := world{db: db, owner: Owner{ID: 1, Email: "sara@example.test", Threshold: schemas.DefaultReviewThreshold}}
	w.team = schemas.Project{Name: "Team", SpaceID: &space.ID}
	w.named = schemas.Project{Name: "Sablier"}
	w.inTeam = schemas.Project{Name: "Sablier", SpaceID: &space.ID}
	w.other = schemas.Project{Name: "Running"}
	w.hidden = schemas.Project{Name: "Hidden", SpaceID: &foreign.ID}
	create(t, db, &w.team, &w.named, &w.inTeam, &w.other, &w.hidden)
	w.opusTask = schemas.Task{ProjectID: w.team.ID, Name: "Spec", FacileID: &teamFacile}
	w.aiTask = schemas.Task{ProjectID: w.team.ID, Name: "Billing"}
	w.runTask = schemas.Task{ProjectID: w.other.ID, Name: "Ops"}
	create(t, db, &w.opusTask, &w.aiTask, &w.runTask)
	stopped := start.Add(time.Hour)
	create(t, db, &schemas.TimeEntry{UserID: 1, ProjectID: w.other.ID, TaskID: w.runTask.ID, StartedAt: start.Add(-time.Hour), StoppedAt: &stopped})
	return w
}

func TestResolveSteps(t *testing.T) {
	w := seedWorld(t)
	aiLabel := "Facile › Team › Billing"
	cases := []struct {
		name        string
		facts       Facts
		decider     *stubDecider
		wantSource  string
		wantProject int64
		wantTask    int64
		wantReview  bool
	}{
		{"explicit task", Facts{TaskID: w.aiTask.ID, ProjectHint: "Sablier"}, nil, SourceExplicit, w.team.ID, w.aiTask.ID, false},
		{"explicit project", Facts{ProjectID: w.other.ID, OpusTaskIDs: []string{"opus-task-1"}}, nil, SourceExplicit, w.other.ID, 0, false},
		{"opus task", Facts{OpusTaskIDs: []string{"unknown", "opus-task-1"}, ProjectHint: "Sablier"}, nil, SourceOpus, w.team.ID, w.opusTask.ID, false},
		{"name inside the registre space", Facts{ProjectHint: "sablier", RegistreSpaceID: "reg-team"}, nil, SourceName, w.inTeam.ID, 0, false},
		{"name across visible projects", Facts{ProjectHint: "SABLIER"}, nil, SourceName, w.named.ID, 0, false},
		{"a foreign project's name is no match", Facts{ProjectHint: "Hidden"}, nil, SourceFallback, 0, 0, true},
		{"running entry", Facts{ProjectHint: "Unknown", At: start}, nil, SourceRunning, w.other.ID, w.runTask.ID, false},
		{"ai pick", Facts{Title: "invoice export", RegistreSpaceID: "reg-team"}, &stubDecider{choice: aiLabel, confidence: 0.92}, SourceAI, w.team.ID, w.aiTask.ID, false},
		{"unsure ai pick needs review", Facts{Title: "invoice export"}, &stubDecider{choice: aiLabel, confidence: 0.6}, SourceAI, w.team.ID, w.aiTask.ID, true},
		{"ai off falls back", Facts{Title: "invoice export"}, &stubDecider{}, SourceFallback, 0, 0, true},
	}
	for _, tc := range cases {
		resolver := Resolver{}
		if tc.decider != nil {
			resolver.Decider = tc.decider
		}
		results, err := resolver.Resolve(context.Background(), w.db, w.owner, []Facts{tc.facts})
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		got := results[0]
		var project, task int64
		if got.Project != nil {
			project = got.Project.ID
		}
		if got.Task != nil {
			task = got.Task.ID
		}
		if got.Source != tc.wantSource || project != tc.wantProject || task != tc.wantTask || got.NeedsReview != tc.wantReview {
			t.Fatalf("%s: got %s project=%d task=%d review=%v", tc.name, got.Source, project, task, got.NeedsReview)
		}
	}

	if _, err := (Resolver{}).Resolve(context.Background(), w.db, w.owner, []Facts{{ProjectID: w.hidden.ID}}); !stderrors.Is(err, ErrNotVisible) {
		t.Fatalf("a project in a foreign space must be refused, got %v", err)
	}
	batch := &stubDecider{choice: aiLabel, confidence: 0.9}
	results, err := Resolver{Decider: batch}.Resolve(context.Background(), w.db, w.owner, []Facts{{Title: "a"}, {Title: "b"}, {ProjectHint: "Sablier"}})
	if err != nil || batch.calls != 1 || results[0].Source != SourceAI || results[1].Source != SourceAI || results[2].Source != SourceName {
		t.Fatalf("one decide call must cover every open item: calls=%d results=%+v err=%v", batch.calls, results, err)
	}
}
