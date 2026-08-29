package spaces

import (
	"context"
	"fmt"
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
