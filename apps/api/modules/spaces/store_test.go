package spaces

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	porteSpaces "github.com/FacileStudio/porte/spaces"
	"github.com/FacileStudio/porte/spaces/spacestest"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestORM(t *testing.T, name string) *gorm.DB {
	t.Helper()

	orm, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := orm.AutoMigrate(&schemas.User{}, &schemas.Space{}, &schemas.SpaceMember{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	return orm
}

// aliasStore runs the conformance suite against the real Store. The suite's
// user ids are opaque strings and space_members.user_id is an int64, so each
// fixture name is given a row id on the way in and translated back on the way
// out. Every read the suite makes still goes through the real SQL.
type aliasStore struct {
	inner  *Store
	orm    *gorm.DB
	byName map[string]int64
	byID   map[int64]string
}

func newAliasStore(orm *gorm.DB) *aliasStore {
	return &aliasStore{
		inner:  NewStore(orm),
		orm:    orm,
		byName: map[string]int64{},
		byID:   map[int64]string{},
	}
}

func (a *aliasStore) id(name string) string {
	if id, ok := a.byName[name]; ok {
		return strconv.FormatInt(id, 10)
	}
	id := int64(len(a.byName) + 1)
	a.byName[name] = id
	a.byID[id] = name
	return strconv.FormatInt(id, 10)
}

func (a *aliasStore) name(id string) string {
	parsed, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return id
	}
	return a.byID[parsed]
}

func (a *aliasStore) translate(member porteSpaces.Membership) porteSpaces.Membership {
	member.UserID = a.name(member.UserID)
	return member
}

func (a *aliasStore) Seed(_ context.Context, member porteSpaces.Membership) error {
	uid, err := strconv.ParseInt(a.id(member.UserID), 10, 64)
	if err != nil {
		return err
	}
	row := schemas.SpaceMember{SpaceID: member.SpaceID, UserID: uid, Role: string(member.Role)}
	return a.orm.Create(&row).Error
}

func (a *aliasStore) Membership(ctx context.Context, spaceID, userID string) (porteSpaces.Membership, error) {
	member, err := a.inner.Membership(ctx, spaceID, a.id(userID))
	if err != nil {
		return porteSpaces.Membership{}, err
	}
	return a.translate(member), nil
}

func (a *aliasStore) Memberships(ctx context.Context, userID string) ([]porteSpaces.Membership, error) {
	rows, err := a.inner.Memberships(ctx, a.id(userID))
	if err != nil {
		return nil, err
	}
	for i, row := range rows {
		rows[i] = a.translate(row)
	}
	return rows, nil
}

func (a *aliasStore) CountRole(ctx context.Context, spaceID string, role porteSpaces.Role) (int, error) {
	return a.inner.CountRole(ctx, spaceID, role)
}

func TestStoreConformance(t *testing.T) {
	instance := 0
	spacestest.Conformance(t, func() porteSpaces.Store {
		instance++
		return newAliasStore(newTestORM(t, fmt.Sprintf("conformance-%d", instance)))
	})
}

// TestStoreConformanceOnPostgres runs the same suite against the database the
// app ships on. SQLite proves the Go mapping and nothing about the SQL: the
// dialects disagree on locking, on comparison and on what a driver hands back
// for an integer column, and only one of them is in production.
func TestStoreConformanceOnPostgres(t *testing.T) {
	url := os.Getenv("SABLIER_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("SABLIER_TEST_DATABASE_URL is unset")
	}
	spacestest.Conformance(t, func() porteSpaces.Store {
		orm, err := newPostgresORM(url)
		if err != nil {
			panic(fmt.Sprintf("open the test database: %v", err))
		}
		return newAliasStore(orm)
	})
}

// TestMembershipReadsEveryFieldFromTheRow bypasses aliasStore, because the
// conformance suite cannot catch a Store that echoes its arguments: its alias
// maps are inverses, so an echoed user id survives the trip back through them
// and lands on the right fixture name.
//
// The user id is asked for as "042" and stored as 42. Both select the same row,
// because the store parses before it queries, so the returned spelling is the
// one thing that tells a read of the row apart from a copy of the argument.
func TestMembershipReadsEveryFieldFromTheRow(t *testing.T) {
	t.Run("sqlite", func(t *testing.T) { assertMembershipIsFaithful(t, newTestORM(t, "faithful")) })
	t.Run("postgres", func(t *testing.T) { assertMembershipIsFaithful(t, openPostgres(t)) })
}

func assertMembershipIsFaithful(t *testing.T, orm *gorm.DB) {
	t.Helper()

	space := schemas.Space{Name: "Shared"}
	if err := orm.Create(&space).Error; err != nil {
		t.Fatalf("create space: %v", err)
	}
	row := schemas.SpaceMember{SpaceID: space.ID, UserID: 42, Role: schemas.RoleAdmin}
	if err := orm.Create(&row).Error; err != nil {
		t.Fatalf("seed member: %v", err)
	}

	got, err := NewStore(orm).Membership(context.Background(), space.ID, "042")
	if err != nil {
		t.Fatalf("Membership: %v", err)
	}
	if got.SpaceID != row.SpaceID {
		t.Fatalf("SpaceID = %q, want %q", got.SpaceID, row.SpaceID)
	}
	if got.UserID != "42" {
		t.Fatalf("UserID = %q, want %q, the row's value and not the argument", got.UserID, "42")
	}
	if got.Role != porteSpaces.RoleAdmin {
		t.Fatalf("Role = %q, want %q", got.Role, porteSpaces.RoleAdmin)
	}
}

func TestMembershipReportsAbsenceAsAnError(t *testing.T) {
	orm := newTestORM(t, "absence")
	store := NewStore(orm)

	if _, err := store.Membership(context.Background(), "no-space", "1"); err != porteSpaces.ErrNotMember {
		t.Fatalf("Membership on an empty table = %v, want ErrNotMember", err)
	}
	if _, err := store.Membership(context.Background(), "no-space", "not-a-number"); err != porteSpaces.ErrNotMember {
		t.Fatalf("Membership on an unparseable user id = %v, want ErrNotMember", err)
	}
}
