package spaces

import (
	"context"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
	portepg "github.com/FacileStudio/porte/pg"

	"gorm.io/gorm"
)

const erasureIssuer = "https://registre.example.test"

func openErasureDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.Migrated(t)
}

func erasureReceiver(t *testing.T, db *gorm.DB) *Deletions {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	accounts, err := auth.NewAccounts(context.Background(), sqlDB, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	return NewDeletions(db, directory.Client{Issuer: erasureIssuer}, accounts.Store().EraseAccount,
		portepg.New(sqlDB), t.TempDir(), slog.New(slog.NewTextHandler(io.Discard, nil)))
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
	mustCreate(t, db, &schemas.Heartbeat{UserID: gone, At: time.Now(), Source: "editor"})
	mustCreate(t, db, &schemas.DraftDismissal{UserID: agent.ID, StartedAt: time.Now(), StoppedAt: time.Now()})
	mustExec(t, db, `INSERT INTO porte_sessions (token_hash, user_id, created_at, last_used_at) VALUES ('h', ?, now(), now())`, gone)
	mustExec(t, db, `INSERT INTO mcp_actions (id, user_id, kind, target, digest, expires_at) VALUES ('a1', ?, 'k', 't', 'd', now())`,
		strconv.FormatInt(gone, 10))
	mustExec(t, db, `INSERT INTO antenne_outbox (channel, payload, idempotency_key) VALUES
		('time_entry.created', '{"payload": {"user_email": "gone@facile.studio"}}', 'k1'),
		('project.created', '{"payload": {"space": {"owner_subject": "sub-gone"}}}', 'k2'),
		('project.created', '{"payload": {"space": {"owner_email": "other@facile.studio"}}}', 'k3')`)
	return erasureFixture{gone: gone, other: other, deletions: erasureReceiver(t, db)}
}

// assertNothingNames fails when any table still holds a row that references
// the erased user or one of their agents.
func assertNothingNames(t *testing.T, db *gorm.DB, gone int64) {
	t.Helper()
	ids := `(SELECT ? UNION SELECT id FROM users WHERE owner_user_id = ?)`
	for _, query := range []string{
		`SELECT count(*) FROM users WHERE id IN ` + ids + ` OR owner_user_id = ?`,
		`SELECT count(*) FROM time_entries WHERE user_id IN ` + ids,
		`SELECT count(*) FROM projects WHERE owner_id IN ` + ids,
		`SELECT count(*) FROM tasks WHERE actor_id IN ` + ids,
		`SELECT count(*) FROM heartbeats WHERE user_id IN ` + ids,
		`SELECT count(*) FROM draft_dismissals WHERE user_id IN ` + ids,
		`SELECT count(*) FROM push_subscriptions WHERE user_id IN ` + ids,
		`SELECT count(*) FROM space_members WHERE user_id IN ` + ids,
		`SELECT count(*) FROM porte_identities WHERE user_id IN ` + ids + ` OR subject = 'sub-gone'`,
		`SELECT count(*) FROM porte_sessions WHERE user_id IN ` + ids,
	} {
		args := []any{gone, gone}
		for range strings.Count(query, "?") - 2 {
			args = append(args, gone)
		}
		if got := count(t, db, query, args...); got != 0 {
			t.Errorf("%s = %d, want 0", query, got)
		}
	}
	if got := count(t, db, `SELECT count(*) FROM antenne_outbox WHERE idempotency_key IN ('k1', 'k2')`); got != 0 {
		t.Errorf("bus events naming the erased person = %d, want 0", got)
	}
	if got := count(t, db, `SELECT count(*) FROM mcp_actions WHERE user_id = ?`, strconv.FormatInt(gone, 10)); got != 0 {
		t.Errorf("pending MCP actions = %d, want 0", got)
	}
}

func TestErasePersonDeletesPersonalDataAndKeepsSpaceData(t *testing.T) {
	assertErasure(t, openErasureDatabase(t))
}

func assertErasure(t *testing.T, db *gorm.DB) {
	t.Helper()
	f := seedErasure(t, db)
	for range 2 {
		if err := f.deletions.ErasePerson(context.Background(), erasureIssuer, "sub-gone"); err != nil {
			t.Fatalf("erase: %v", err)
		}
		if err := f.deletions.receiver.Erase(context.Background(), erasureIssuer, "sub-gone"); err != nil {
			t.Fatalf("erase the identity: %v", err)
		}
	}
	assertNothingNames(t, db, f.gone)

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
	if n := count(t, db, `SELECT count(*) FROM antenne_outbox WHERE idempotency_key = 'k3'`); n != 1 {
		t.Errorf("another person's bus event was erased")
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
	deletions := erasureReceiver(t, db)
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

func TestEraseIdentityToleratesTheSchemaWithoutMagicLinks(t *testing.T) {
	db := openPostgres(t)
	if err := schemas.AdoptPorte(db, ""); err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := portepg.New(sqlDB).EraseIdentity(context.Background(), erasureIssuer, "sub-nobody"); err != nil {
		t.Fatalf("EraseIdentity on Sablier's schema: %v", err)
	}
}
