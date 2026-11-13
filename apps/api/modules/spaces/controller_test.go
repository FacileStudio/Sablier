package spaces

import (
	"context"
	"net/http"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

func seedMember(t *testing.T, orm *gorm.DB, spaceID string, userID int64, role string) {
	t.Helper()

	member := schemas.SpaceMember{SpaceID: spaceID, UserID: userID, Role: role}
	if err := orm.Create(&member).Error; err != nil {
		t.Fatalf("seed member: %v", err)
	}
}

func newTestSpace(t *testing.T) (*Service, string) {
	t.Helper()

	orm := testdb.Migrated(t)
	space := schemas.Space{Name: "Shared"}
	if err := orm.Create(&space).Error; err != nil {
		t.Fatalf("create space: %v", err)
	}
	seedMember(t, orm, space.ID, 1, schemas.RoleOwner)
	seedMember(t, orm, space.ID, 2, schemas.RoleAdmin)
	seedMember(t, orm, space.ID, 3, schemas.RoleMember)
	return NewService(orm), space.ID
}

func assertStatus(t *testing.T, err error, want int, message string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected an error with status %d, got none", want)
	}
	if got := errors.Status(err); got != want {
		t.Fatalf("status = %d, want %d (%v)", got, want, err)
	}
	if err.Error() != message {
		t.Fatalf("message = %q, want %q", err.Error(), message)
	}
}

func TestGetRefusesANonMember(t *testing.T) {
	service, spaceID := newTestSpace(t)

	_, err := service.controller.get(context.Background(), spaceID, "99")
	assertStatus(t, err, http.StatusForbidden, "not a member of this space")
}

func TestListDropsAMembershipTheLadderDoesNotRank(t *testing.T) {
	service, spaceID := newTestSpace(t)
	other := schemas.Space{Name: "Corrupt"}
	if err := service.orm.Create(&other).Error; err != nil {
		t.Fatalf("create space: %v", err)
	}
	seedMember(t, service.orm, other.ID, 3, "superuser")

	list, err := service.controller.list(context.Background(), "3")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list.Spaces) != 1 || list.Spaces[0].ID != spaceID {
		t.Fatalf("spaces = %+v, want only the ranked membership in %s", list.Spaces, spaceID)
	}
}

// Every package in this module shares one SABLIER_TEST_DATABASE_URL and `go
// test ./...` runs them at the same time, so this one owns a schema of its own
// rather than dropping `public` out from under a neighbour mid-run.
