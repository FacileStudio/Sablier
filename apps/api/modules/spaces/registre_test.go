package spaces

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"

	"gorm.io/gorm"
)

const registreIssuer = "https://sso.example.test"

func newRegistre(orm *gorm.DB) *Registre {
	return NewRegistre(orm, registreIssuer, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func mustExec(t *testing.T, orm *gorm.DB, statement string, args ...any) {
	t.Helper()
	if err := orm.Exec(statement, args...).Error; err != nil {
		t.Fatalf("seed: %v\n%s", err, statement)
	}
}

// seedTeam builds one unlinked space, "team", holding: alice (Registre identity, owner),
// carol (local password account, member), dave (Registre identity, admin, about to be
// removed in Registre) and an agent user. bob exists with no identity and no membership.
func seedTeam(t *testing.T, orm *gorm.DB) string {
	t.Helper()
	ids := map[string]int64{}
	for _, name := range []string{"alice", "bob", "carol", "dave"} {
		user := schemas.User{Email: name + "@x.test"}
		if err := orm.Create(&user).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
		ids[name] = user.ID
	}
	mustExec(t, orm, `INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, 'sub-alice'), (?, ?, 'sub-dave')`,
		ids["alice"], registreIssuer, ids["dave"], registreIssuer)
	space := schemas.Space{Name: "team"}
	if err := orm.Create(&space).Error; err != nil {
		t.Fatalf("seed space: %v", err)
	}
	seedMember(t, orm, space.ID, ids["alice"], schemas.RoleOwner)
	seedMember(t, orm, space.ID, ids["carol"], schemas.RoleMember)
	seedMember(t, orm, space.ID, ids["dave"], schemas.RoleAdmin)
	agent := schemas.User{Email: "agent@agents.invalid", Kind: schemas.UserKindAgent}
	if err := orm.Create(&agent).Error; err != nil {
		t.Fatalf("seed agent: %v", err)
	}
	seedMember(t, orm, space.ID, agent.ID, schemas.RoleMember)
	return space.ID
}

func teamSnapshot() directory.Snapshot {
	return directory.Snapshot{Version: 7, Spaces: []directory.Space{
		{ID: "r-team", Name: "Team", Members: []directory.Member{
			{Subject: "sub-alice", Email: "alice@x.test", Role: directory.RoleOwner},
			{Subject: "sub-bob", Email: "BOB@x.test", Role: directory.RoleAdmin},
			{Subject: "sub-erin", Email: "erin@x.test", Name: "Erin", Picture: registreIssuer + "/avatars/sub-erin", Role: "viewer"},
		}},
		{ID: "r-fresh", Name: "Fresh", Members: []directory.Member{
			{Subject: "sub-alice", Email: "alice@x.test", Role: directory.RoleMember},
		}},
	}}
}

// membershipState is every row Apply can touch, flattened so two applies can be compared.
func membershipState(t *testing.T, orm *gorm.DB) map[string]string {
	t.Helper()
	var rows []struct{ ID, SpaceID, Email, Role, Link string }
	err := orm.Raw(`SELECT m.id, m.space_id, u.email, m.role, COALESCE(s.registre_space_id, '') || ':' || s.name AS link
		FROM space_members m JOIN users u ON u.id = m.user_id JOIN spaces s ON s.id = m.space_id`).Scan(&rows).Error
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	state := map[string]string{}
	for _, row := range rows {
		state[row.Link+"/"+row.Email] = row.Role + "#" + row.ID
	}
	return state
}

func roleOf(state map[string]string, key string) string {
	role, _, _ := strings.Cut(state[key], "#")
	return role
}

func TestApplyMirrorsTheSnapshot(t *testing.T) {
	assertApply(t, testdb.Migrated(t))
}

func assertApply(t *testing.T, orm *gorm.DB) {
	t.Helper()
	localID := seedTeam(t, orm)
	registre := newRegistre(orm)
	if err := registre.Apply(context.Background(), teamSnapshot()); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	var linked schemas.Space
	if err := orm.First(&linked, "id = ?", localID).Error; err != nil || linked.RegistreSpaceID == nil || *linked.RegistreSpaceID != "r-team" {
		t.Fatalf("local team = %+v (%v), want it linked to r-team by name", linked, err)
	}
	state := membershipState(t, orm)
	want := map[string]string{
		"r-team:Team/alice@x.test":         schemas.RoleOwner,
		"r-team:Team/bob@x.test":           schemas.RoleAdmin,
		"r-team:Team/erin@x.test":          schemas.RoleMember,
		"r-team:Team/carol@x.test":         schemas.RoleMember,
		"r-team:Team/agent@agents.invalid": schemas.RoleMember,
		"r-fresh:Fresh/alice@x.test":       schemas.RoleMember,
	}
	if len(state) != len(want) {
		t.Fatalf("memberships = %v, want exactly %v (dave removed, carol and the agent kept)", state, want)
	}
	for key, role := range want {
		if got := roleOf(state, key); got != role {
			t.Errorf("%s role = %q, want %q", key, got, role)
		}
	}
	assertProvisioned(t, orm)

	if err := registre.Apply(context.Background(), teamSnapshot()); err != nil {
		t.Fatalf("second Apply: %v", err)
	}
	again := membershipState(t, orm)
	for key, value := range state {
		if again[key] != value {
			t.Errorf("second apply changed %s: %q -> %q", key, value, again[key])
		}
	}
	var users int64
	orm.Model(&schemas.User{}).Count(&users)
	if users != 6 {
		t.Errorf("users after two applies = %d, want 6 (erin provisioned once)", users)
	}
}

func assertProvisioned(t *testing.T, orm *gorm.DB) {
	t.Helper()
	var erin schemas.User
	if err := orm.First(&erin, "email = ?", "erin@x.test").Error; err != nil {
		t.Fatalf("erin was not provisioned: %v", err)
	}
	if erin.Name != "Erin" || erin.OIDCPictureURL != registreIssuer+"/avatars/sub-erin" || erin.Color == "" {
		t.Errorf("erin = %+v, want Registre's name and picture and a colour", erin)
	}
	for subject, email := range map[string]string{"sub-erin": "erin@x.test", "sub-bob": "bob@x.test"} {
		var owner []string
		orm.Raw(`SELECT u.email FROM porte_identities i JOIN users u ON u.id = i.user_id
			WHERE i.provider = ? AND i.subject = ?`, registreIssuer, subject).Scan(&owner)
		if len(owner) != 1 || owner[0] != email {
			t.Errorf("identity %s belongs to %v, want %s", subject, owner, email)
		}
	}
}
