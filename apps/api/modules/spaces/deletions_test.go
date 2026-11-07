package spaces

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const erasureIssuer = "https://registre.example.test"

func openErasureDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&schemas.User{}, &schemas.Project{}, &schemas.Task{}, &schemas.TimeEntry{},
		&schemas.Heartbeat{}, &schemas.DraftDismissal{}, &schemas.PushSubscription{}, &schemas.Space{}, &schemas.SpaceMember{})
	if err != nil {
		t.Fatal(err)
	}
	for _, ddl := range []string{
		`CREATE TABLE porte_identities (user_id integer NOT NULL, provider text NOT NULL, subject text NOT NULL, PRIMARY KEY (provider, subject))`,
		`CREATE TABLE directory_deletion_state (id integer PRIMARY KEY, state blob NOT NULL, updated_at datetime)`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func count(t *testing.T, db *gorm.DB, query string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := db.Raw(query, args...).Scan(&n).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func mustCreate(t *testing.T, db *gorm.DB, row any) {
	t.Helper()
	if err := db.Create(row).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
}

func person(t *testing.T, db *gorm.DB, name string) int64 {
	t.Helper()
	user := schemas.User{Email: name + "@facile.studio", Name: name}
	mustCreate(t, db, &user)
	mustExec(t, db, `INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)`, user.ID, erasureIssuer, "sub-"+name)
	return user.ID
}

type erasureFixture struct {
	gone, other int64
	deletions   *Deletions
}

func seedErasure(t *testing.T, db *gorm.DB) erasureFixture {
	t.Helper()
	gone, other := person(t, db, "gone"), person(t, db, "other")
	agent, err := schemas.EnsureAgentUser(db, gone, schemas.AgentFromName("claude"))
	if err != nil {
		t.Fatal(err)
	}
	space := schemas.Space{Name: "Studio"}
	mustCreate(t, db, &space)
	seedMember(t, db, space.ID, gone, schemas.RoleMember)
	seedMember(t, db, space.ID, other, schemas.RoleOwner)

	mine := schemas.Project{Name: "Perso", OwnerID: gone}
	theirs := schemas.Project{Name: "Autre", OwnerID: other}
	shared := schemas.Project{Name: "Equipe", OwnerID: gone, SpaceID: &space.ID}
	for _, project := range []*schemas.Project{&mine, &theirs, &shared} {
		mustCreate(t, db, project)
	}
	mineTask := schemas.Task{ProjectID: mine.ID, Name: "t1"}
	sharedTask := schemas.Task{ProjectID: shared.ID, Name: "t2", ActorID: &gone}
	theirTask := schemas.Task{ProjectID: theirs.ID, Name: "t3", ActorID: &other}
	for _, task := range []*schemas.Task{&mineTask, &sharedTask, &theirTask} {
		mustCreate(t, db, task)
	}
	for _, entry := range []schemas.TimeEntry{
		{ProjectID: mine.ID, TaskID: mineTask.ID, UserID: gone},
		{ProjectID: shared.ID, TaskID: sharedTask.ID, UserID: gone},
		{ProjectID: shared.ID, TaskID: sharedTask.ID, UserID: agent.ID},
		{ProjectID: shared.ID, TaskID: sharedTask.ID, UserID: other},
		{ProjectID: theirs.ID, TaskID: theirTask.ID, UserID: other},
	} {
		mustCreate(t, db, &entry)
	}
	mustCreate(t, db, &schemas.PushSubscription{UserID: gone, Endpoint: "e", P256DH: "p", Auth: "a"})
	mustCreate(t, db, &schemas.PushSubscription{UserID: other, Endpoint: "e2", P256DH: "p", Auth: "a"})
	deletions := NewDeletions(db, directory.Client{Issuer: erasureIssuer}, nil, t.TempDir(),
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	return erasureFixture{gone: gone, other: other, deletions: deletions}
}

func TestErasePersonDeletesPersonalDataAndKeepsSpaceData(t *testing.T) {
	assertErasure(t, openErasureDatabase(t))
}

func TestErasePersonOnPostgres(t *testing.T) {
	db := openPostgres(t)
	err := db.AutoMigrate(&schemas.Project{}, &schemas.Task{}, &schemas.TimeEntry{}, &schemas.Heartbeat{},
		&schemas.DraftDismissal{}, &schemas.PushSubscription{})
	if err != nil {
		t.Fatal(err)
	}
	if err := schemas.AdoptPorte(db, ""); err != nil {
		t.Fatal(err)
	}
	assertErasure(t, db)
}

func assertErasure(t *testing.T, db *gorm.DB) {
	t.Helper()
	f := seedErasure(t, db)
	for range 2 {
		if err := f.deletions.ErasePerson(context.Background(), erasureIssuer, "sub-gone"); err != nil {
			t.Fatalf("erase: %v", err)
		}
	}

	for query, want := range map[string]int64{
		`SELECT count(*) FROM users WHERE id = ? OR owner_user_id = ?`:         0,
		`SELECT count(*) FROM time_entries WHERE user_id = ?`:                  0,
		`SELECT count(*) FROM projects WHERE name = 'Perso'`:                   0,
		`SELECT count(*) FROM tasks WHERE name = 't1'`:                         0,
		`SELECT count(*) FROM push_subscriptions WHERE user_id = ?`:            0,
		`SELECT count(*) FROM space_members WHERE user_id = ?`:                 0,
		`SELECT count(*) FROM projects WHERE name = 'Equipe' AND owner_id = 0`: 0,
		`SELECT count(*) FROM tasks WHERE name = 't2' AND actor_id IS NULL`:    1,
	} {
		args := []any{f.gone}
		if strings.Count(query, "?") == 2 {
			args = append(args, f.gone)
		} else if !strings.Contains(query, "?") {
			args = nil
		}
		if got := count(t, db, query, args...); got != want {
			t.Errorf("%s = %d, want %d", query, got, want)
		}
	}
	if n := count(t, db, `SELECT count(*) FROM projects WHERE name = 'Equipe' AND owner_id = ?`, f.other); n != 1 {
		t.Errorf("the space project did not pass to the space owner")
	}
	if n := count(t, db, `SELECT count(*) FROM users WHERE id = ?`, f.other); n != 1 {
		t.Errorf("another account was touched")
	}
	if n := count(t, db, `SELECT count(*) FROM time_entries`); n != 2 {
		t.Errorf("other people's time entries = %d, want 2", n)
	}
	if n := count(t, db, `SELECT count(*) FROM projects WHERE name = 'Autre'`); n != 1 {
		t.Errorf("another person's project was touched")
	}
	if n := count(t, db, `SELECT count(*) FROM push_subscriptions WHERE user_id = ?`, f.other); n != 1 {
		t.Errorf("another person's subscription was touched")
	}
}

func TestDeletionStateSurvivesARestart(t *testing.T) {
	db := openErasureDatabase(t)
	state := deletionState{orm: db}

	if saved, err := state.Load(context.Background()); err != nil || saved != nil {
		t.Fatalf("empty Load = %q, %v; want nil", saved, err)
	}
	for _, blob := range []string{"one", "two"} {
		if err := state.Save(context.Background(), []byte(blob)); err != nil {
			t.Fatalf("save: %v", err)
		}
	}
	if saved, err := state.Load(context.Background()); err != nil || string(saved) != "two" {
		t.Fatalf("Load = %q, %v; want two", saved, err)
	}
}

func TestErasePersonTransfersSpaceProjectsToTheLongestStandingAdmin(t *testing.T) {
	db := openErasureDatabase(t)
	gone := person(t, db, "gone")
	oldMember, youngAdmin, oldAdmin := person(t, db, "oldmember"), person(t, db, "youngadmin"), person(t, db, "oldadmin")
	robot := schemas.User{Email: "robot@facile.studio", Kind: schemas.UserKindAgent}
	mustCreate(t, db, &robot)

	crowded, soloMembers, empty := schemas.Space{Name: "Crowded"}, schemas.Space{Name: "Solo"}, schemas.Space{Name: "Empty"}
	for _, space := range []*schemas.Space{&crowded, &soloMembers, &empty} {
		mustCreate(t, db, space)
	}
	seed := func(space schemas.Space, user int64, role, joined string) {
		seedMember(t, db, space.ID, user, role)
		mustExec(t, db, `UPDATE space_members SET joined_at = ? WHERE space_id = ? AND user_id = ?`, joined, space.ID, user)
	}
	seed(crowded, gone, schemas.RoleAdmin, "2020-01-01 00:00:00")
	seed(crowded, robot.ID, schemas.RoleAdmin, "2020-01-02 00:00:00")
	seed(crowded, oldMember, schemas.RoleMember, "2021-01-01 00:00:00")
	seed(crowded, youngAdmin, schemas.RoleAdmin, "2024-01-01 00:00:00")
	seed(crowded, oldAdmin, schemas.RoleAdmin, "2022-01-01 00:00:00")
	seed(soloMembers, gone, schemas.RoleAdmin, "2020-01-01 00:00:00")
	seed(soloMembers, youngAdmin, schemas.RoleMember, "2024-01-01 00:00:00")
	seed(soloMembers, oldMember, schemas.RoleMember, "2021-01-01 00:00:00")
	seed(empty, gone, schemas.RoleOwner, "2020-01-01 00:00:00")

	for name, space := range map[string]schemas.Space{"P-crowded": crowded, "P-solo": soloMembers, "P-empty": empty} {
		mustCreate(t, db, &schemas.Project{Name: name, OwnerID: gone, SpaceID: &space.ID})
	}
	deletions := NewDeletions(db, directory.Client{Issuer: erasureIssuer}, nil, t.TempDir(),
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	for range 2 {
		if err := deletions.ErasePerson(context.Background(), erasureIssuer, "sub-gone"); err != nil {
			t.Fatalf("erase: %v", err)
		}
	}

	for name, want := range map[string]int64{"P-crowded": oldAdmin, "P-solo": oldMember, "P-empty": 0} {
		if got := count(t, db, `SELECT owner_id FROM projects WHERE name = ?`, name); got != want {
			t.Errorf("%s owner = %d, want %d", name, got, want)
		}
	}
}
