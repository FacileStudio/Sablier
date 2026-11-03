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
