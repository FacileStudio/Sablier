package spaces

import (
	"context"
	"strconv"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	porteSpaces "github.com/FacileStudio/porte/spaces"
	"github.com/FacileStudio/porte/spaces/spacestest"

	"gorm.io/gorm"
)

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

// TestStoreConformance runs porte's spaces conformance suite against the real
// Store on the migrated PostgreSQL schema, emptied for every case.
func TestStoreConformance(t *testing.T) {
	testdb.Migrated(t)
	spacestest.Conformance(t, func() porteSpaces.Store {
		return newAliasStore(testdb.Migrated(t))
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
	assertMembershipIsFaithful(t, testdb.Migrated(t))
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
	orm := testdb.Migrated(t)
	store := NewStore(orm)

	if _, err := store.Membership(context.Background(), "no-space", "1"); err != porteSpaces.ErrNotMember {
		t.Fatalf("Membership on an empty table = %v, want ErrNotMember", err)
	}
	if _, err := store.Membership(context.Background(), "no-space", "not-a-number"); err != porteSpaces.ErrNotMember {
		t.Fatalf("Membership on an unparseable user id = %v, want ErrNotMember", err)
	}
}
