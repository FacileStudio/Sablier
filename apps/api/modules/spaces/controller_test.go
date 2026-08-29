package spaces

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func seedMember(t *testing.T, orm *gorm.DB, spaceID string, userID int64, role string) {
	t.Helper()

	member := schemas.SpaceMember{SpaceID: spaceID, UserID: userID, Role: role}
	if err := orm.Create(&member).Error; err != nil {
		t.Fatalf("seed member: %v", err)
	}
}

func newTestSpace(t *testing.T, name string) (*Service, string) {
	t.Helper()

	orm := newTestORM(t, name)
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
	service, spaceID := newTestSpace(t, "guard-get")

	_, err := service.controller.get(context.Background(), spaceID, "99")
	assertStatus(t, err, http.StatusForbidden, "not a member of this space")
}

func TestUpdateRequiresAdmin(t *testing.T) {
	service, spaceID := newTestSpace(t, "guard-update")
	ctx := context.Background()
	req := &UpdateSpaceRequest{Name: "Renamed"}

	_, err := service.controller.update(ctx, spaceID, "3", req)
	assertStatus(t, err, http.StatusForbidden, "only owners and admins can update a space")

	if _, err := service.controller.update(ctx, spaceID, "2", req); err != nil {
		t.Fatalf("admin update: %v", err)
	}
}

func TestDeleteRequiresOwner(t *testing.T) {
	service, spaceID := newTestSpace(t, "guard-delete")

	err := service.controller.delete(context.Background(), spaceID, "2")
	assertStatus(t, err, http.StatusForbidden, "only owners can delete a space")
}

func TestAddMemberRefusesAnAdminMintingAnOwner(t *testing.T) {
	service, spaceID := newTestSpace(t, "guard-add")
	ctx := context.Background()

	_, err := service.controller.addMember(ctx, spaceID, "2", &AddMemberRequest{UserID: 4, Role: schemas.RoleOwner})
	assertStatus(t, err, http.StatusForbidden, "only owners can assign the owner role")

	user := schemas.User{ID: 4, Email: "four@example.test"}
	if err := service.orm.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	_, err = service.controller.addMember(ctx, spaceID, "2", &AddMemberRequest{UserID: 4, Role: schemas.RoleAdmin})
	if err != nil {
		t.Fatalf("admin adding a peer admin: %v", err)
	}
}

func TestLeaveRefusesTheSoleOwnerAndAllowsAPeer(t *testing.T) {
	service, spaceID := newTestSpace(t, "guard-leave")
	ctx := context.Background()

	err := service.controller.leave(ctx, spaceID, "1")
	assertStatus(t, err, http.StatusPreconditionFailed,
		"cannot leave space as the only owner; transfer ownership or delete the space")

	if err := service.controller.leave(ctx, spaceID, "3"); err != nil {
		t.Fatalf("member leaving: %v", err)
	}

	seedMember(t, service.orm, spaceID, 5, schemas.RoleOwner)
	if err := service.controller.leave(ctx, spaceID, "1"); err != nil {
		t.Fatalf("owner with a peer leaving: %v", err)
	}
}

func TestLeaveRefusesANonMember(t *testing.T) {
	service, spaceID := newTestSpace(t, "guard-leave-stranger")

	err := service.controller.leave(context.Background(), spaceID, "99")
	assertStatus(t, err, http.StatusForbidden, "not a member of this space")
}

// The row lock leaveSpace takes is Postgres-only SQL, so SQLite proves nothing
// about it: a malformed locking clause would surface as a 500 in production and
// nowhere else.
func openPostgres(t *testing.T) *gorm.DB {
	t.Helper()

	url := os.Getenv("SABLIER_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("SABLIER_TEST_DATABASE_URL is unset")
	}
	orm, err := gorm.Open(postgres.Open(url), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := orm.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public").Error; err != nil {
		t.Fatalf("reset the schema: %v", err)
	}
	if err := orm.AutoMigrate(&schemas.User{}, &schemas.Space{}, &schemas.SpaceMember{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return orm
}

func TestLeaveLocksTheMemberRowsOnPostgres(t *testing.T) {
	orm := openPostgres(t)
	space := schemas.Space{Name: "Shared"}
	if err := orm.Create(&space).Error; err != nil {
		t.Fatalf("create space: %v", err)
	}
	seedMember(t, orm, space.ID, 1, schemas.RoleOwner)
	seedMember(t, orm, space.ID, 2, schemas.RoleOwner)
	service := NewService(orm)

	if err := service.controller.leave(context.Background(), space.ID, "2"); err != nil {
		t.Fatalf("owner with a peer leaving: %v", err)
	}
	err := service.controller.leave(context.Background(), space.ID, "1")
	assertStatus(t, err, http.StatusPreconditionFailed,
		"cannot leave space as the only owner; transfer ownership or delete the space")
}
