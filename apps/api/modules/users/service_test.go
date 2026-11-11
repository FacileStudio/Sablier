package users

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

func TestGetUserAssignsLeastUsedColorWhenMissing(t *testing.T) {
	service := newDatabaseBackedService(t)
	seedUser(t, service.orm, "alpha@example.com", "AD9EF0")
	missing := seedUser(t, service.orm, "beta@example.com", "")

	user, err := service.getUser(context.Background(), fmt.Sprintf("%d", missing.ID))
	if err != nil {
		t.Fatalf("get user: %v", err)
	}

	if user.Color != "F09ED6" {
		t.Fatalf("expected F09ED6, got %s", user.Color)
	}

	var record schemas.User
	if err := service.orm.Where("id = ?", missing.ID).First(&record).Error; err != nil {
		t.Fatalf("read updated user: %v", err)
	}
	if record.Color != "F09ED6" {
		t.Fatalf("expected stored color F09ED6, got %s", record.Color)
	}
}

func TestUpdateUserPersistsChosenColor(t *testing.T) {
	service := newDatabaseBackedService(t)
	user := seedUser(t, service.orm, "color@example.com", "AD9EF0")
	color := "7EEEDB"

	updated, err := service.updateUser(context.Background(), fmt.Sprintf("%d", user.ID), nil, &color, nil, nil, nil)
	if err != nil {
		t.Fatalf("update user: %v", err)
	}

	if updated.Color != color {
		t.Fatalf("expected color %s, got %s", color, updated.Color)
	}
}

func TestListUsersBackfillsMissingColors(t *testing.T) {
	service := newDatabaseBackedService(t)
	one := seedUser(t, service.orm, "one@example.com", "")
	two := seedUser(t, service.orm, "two@example.com", "")
	shareSpace(t, service.orm, one, two)

	users, err := service.listUsers(context.Background(), "1")
	if err != nil {
		t.Fatalf("list users: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Color == "" || users[1].Color == "" {
		t.Fatalf("expected colors assigned, got %#v", users)
	}
	if users[0].Color == users[1].Color {
		t.Fatalf("expected distributed colors, got %#v", users)
	}
}

func TestUpdateMeRejectsInvalidColor(t *testing.T) {
	controller := &Controller{}
	color := "FFFFFF"
	req := &UpdateRequest{Color: &color}
	request := httptest.NewRequest(http.MethodPatch, "/api/users/me", nil)
	ctx := authcontext.WithIdentity(request.Context(), authcontext.Identity{
		UserID: "1",
		Email:  "user@example.com",
	})

	_, err := controller.updateMe(request.WithContext(ctx), req)
	if err == nil || err.Error() != "color must be one of: AD9EF0, F09ED6, EE7E89, EEB47E, A9EE7E, 7EEEDB" {
		t.Fatalf("expected invalid color error, got %v", err)
	}
}

func TestListUsersShowsOnlySelfOwnAgentsAndSpacePeers(t *testing.T) {
	service := newDatabaseBackedService(t)
	caller := seedUser(t, service.orm, "caller@example.com", "AD9EF0")
	peer := seedUser(t, service.orm, "peer@example.com", "F09ED6")
	stranger := seedUser(t, service.orm, "stranger@example.com", "EE7E89")
	agent := schemas.User{Email: "agent@example.com", Color: "EEB47E", Kind: schemas.UserKindAgent, OwnerUserID: &caller.ID}
	if err := service.orm.Create(&agent).Error; err != nil {
		t.Fatalf("create agent: %v", err)
	}
	shareSpace(t, service.orm, caller, peer)
	shareSpace(t, service.orm, stranger)

	users, err := service.listUsers(context.Background(), fmt.Sprint(caller.ID))
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	got := map[string]bool{}
	for _, user := range users {
		got[user.Email] = true
	}
	if len(got) != 3 || !got["caller@example.com"] || !got["peer@example.com"] || !got["agent@example.com"] {
		t.Fatalf("visible users = %v, want the caller, its agent and its space peer", got)
	}

	if _, err := service.getPeer(context.Background(), fmt.Sprint(caller.ID), fmt.Sprint(peer.ID)); err != nil {
		t.Fatalf("get a space peer: %v", err)
	}
	_, err = service.getPeer(context.Background(), fmt.Sprint(caller.ID), fmt.Sprint(stranger.ID))
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) || apiErr.Code != "not_found" {
		t.Fatalf("get a stranger = %v, want not_found", err)
	}
}

func TestUpdateMeIgnoresEmail(t *testing.T) {
	service := newDatabaseBackedService(t)
	user := seedUser(t, service.orm, "kept@example.com", "AD9EF0")
	request := httptest.NewRequest(http.MethodPatch, "/api/users/me", nil)
	ctx := authcontext.WithIdentity(request.Context(), authcontext.Identity{UserID: fmt.Sprint(user.ID), Email: user.Email})
	var body UpdateRequest
	if err := json.Unmarshal([]byte(`{"email":"admin@example.com","color":"7EEEDB"}`), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	updated, err := service.controller.updateMe(request.WithContext(ctx), &body)
	if err != nil {
		t.Fatalf("update me: %v", err)
	}
	if updated.User.Email != "kept@example.com" {
		t.Fatalf("email = %q, want it unchanged", updated.User.Email)
	}
}

func shareSpace(t *testing.T, orm *gorm.DB, members ...schemas.User) {
	t.Helper()
	space := schemas.Space{Name: "shared"}
	if err := orm.Create(&space).Error; err != nil {
		t.Fatalf("create space: %v", err)
	}
	for _, member := range members {
		if err := orm.Create(&schemas.SpaceMember{SpaceID: space.ID, UserID: member.ID, Role: schemas.RoleMember}).Error; err != nil {
			t.Fatalf("add member: %v", err)
		}
	}
}

func newDatabaseBackedService(t *testing.T) *Service {
	t.Helper()

	return NewService(testdb.Open(t), nil, nil)
}

func seedUser(t *testing.T, orm *gorm.DB, email string, color string) schemas.User {
	t.Helper()

	user := schemas.User{
		Email:        email,
		Color:        color,
		PasswordHash: "hash",
	}
	if err := orm.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}
