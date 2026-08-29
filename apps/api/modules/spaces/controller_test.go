package spaces

import (
	"context"
	"net/http"
	"net/url"
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

func ownerMemberID(t *testing.T, service *Service, spaceID string, userID int64) string {
	t.Helper()

	var member schemas.SpaceMember
	err := service.orm.Where("space_id = ? AND user_id = ?", spaceID, userID).First(&member).Error
	if err != nil {
		t.Fatalf("read member row: %v", err)
	}
	return member.ID
}

func countOwners(t *testing.T, service *Service, spaceID string) int64 {
	t.Helper()

	var count int64
	err := service.orm.Model(&schemas.SpaceMember{}).
		Where("space_id = ? AND role = ?", spaceID, schemas.RoleOwner).
		Count(&count).Error
	if err != nil {
		t.Fatalf("count owners: %v", err)
	}
	return count
}

func TestRemoveMemberRefusesTheSoleOwner(t *testing.T) {
	service, spaceID := newTestSpace(t, "guard-remove-owner")
	owner := ownerMemberID(t, service, spaceID, 1)

	err := service.controller.removeMember(context.Background(), spaceID, "1", owner)
	assertStatus(t, err, http.StatusPreconditionFailed,
		"cannot remove the last owner; promote another owner first")

	if got := countOwners(t, service, spaceID); got != 1 {
		t.Fatalf("owners after the refusal = %d, want 1", got)
	}
}

func TestRemoveMemberRefusesAnAdminRemovingAnOwner(t *testing.T) {
	service, spaceID := newTestSpace(t, "guard-remove-outranks")
	seedMember(t, service.orm, spaceID, 5, schemas.RoleOwner)
	owner := ownerMemberID(t, service, spaceID, 1)

	err := service.controller.removeMember(context.Background(), spaceID, "2", owner)
	assertStatus(t, err, http.StatusForbidden, "cannot remove a member who outranks you")

	if err := service.controller.removeMember(context.Background(), spaceID, "1", owner); err != nil {
		t.Fatalf("owner removing a peer owner: %v", err)
	}
}

func TestRemoveMemberAllowsAnAdminRemovingAPeer(t *testing.T) {
	service, spaceID := newTestSpace(t, "guard-remove-peer")
	peer := ownerMemberID(t, service, spaceID, 3)

	if err := service.controller.removeMember(context.Background(), spaceID, "2", peer); err != nil {
		t.Fatalf("admin removing a member: %v", err)
	}
}

func TestUpdateMemberRoleRefusesDemotingTheSoleOwner(t *testing.T) {
	service, spaceID := newTestSpace(t, "guard-demote-owner")
	owner := ownerMemberID(t, service, spaceID, 1)
	req := &UpdateMemberRoleRequest{Role: schemas.RoleMember}

	_, err := service.controller.updateMemberRole(context.Background(), spaceID, "1", owner, req)
	assertStatus(t, err, http.StatusPreconditionFailed,
		"cannot remove the last owner; promote another owner first")

	if got := countOwners(t, service, spaceID); got != 1 {
		t.Fatalf("owners after the refusal = %d, want 1", got)
	}
}

func TestUpdateMemberRoleAllowsDemotingAnOwnerWithAPeer(t *testing.T) {
	service, spaceID := newTestSpace(t, "guard-demote-peer")
	seedMember(t, service.orm, spaceID, 5, schemas.RoleOwner)
	owner := ownerMemberID(t, service, spaceID, 1)

	req := &UpdateMemberRoleRequest{Role: schemas.RoleAdmin}
	if _, err := service.controller.updateMemberRole(context.Background(), spaceID, "5", owner, req); err != nil {
		t.Fatalf("owner demoting a peer owner: %v", err)
	}
	if got := countOwners(t, service, spaceID); got != 1 {
		t.Fatalf("owners after the demotion = %d, want 1", got)
	}
}

func TestListDropsAMembershipTheLadderDoesNotRank(t *testing.T) {
	service, spaceID := newTestSpace(t, "guard-list")
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

func TestLeaveRefusesANonMember(t *testing.T) {
	service, spaceID := newTestSpace(t, "guard-leave-stranger")

	err := service.controller.leave(context.Background(), spaceID, "99")
	assertStatus(t, err, http.StatusForbidden, "not a member of this space")
}

// The row lock leaveSpace takes is Postgres-only SQL, so SQLite proves nothing
// about it: a malformed locking clause would surface as a 500 in production and
// nowhere else.
//
// Every package in this module shares one SABLIER_TEST_DATABASE_URL and `go
// test ./...` runs them at the same time, so this one owns a schema of its own
// rather than dropping `public` out from under a neighbour mid-run.
const testSchema = "spaces_test"

func scopedURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set("search_path", testSchema)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func newPostgresORM(raw string) (*gorm.DB, error) {
	bootstrap, err := gorm.Open(postgres.Open(raw), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return nil, err
	}
	err = bootstrap.Exec("DROP SCHEMA IF EXISTS " + testSchema + " CASCADE; CREATE SCHEMA " + testSchema).Error
	if err != nil {
		return nil, err
	}
	pool, err := bootstrap.DB()
	if err != nil {
		return nil, err
	}
	if err := pool.Close(); err != nil {
		return nil, err
	}

	scoped, err := scopedURL(raw)
	if err != nil {
		return nil, err
	}
	orm, err := gorm.Open(postgres.Open(scoped), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return nil, err
	}
	if err := orm.AutoMigrate(&schemas.User{}, &schemas.Space{}, &schemas.SpaceMember{}); err != nil {
		return nil, err
	}
	return orm, nil
}

func openPostgres(t *testing.T) *gorm.DB {
	t.Helper()

	url := os.Getenv("SABLIER_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("SABLIER_TEST_DATABASE_URL is unset")
	}
	orm, err := newPostgresORM(url)
	if err != nil {
		t.Fatalf("open the test database: %v", err)
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
