package auth

import (
	"context"
	stderrors "errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAgentUserCannotSignIn(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	orm, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := orm.AutoMigrate(&schemas.User{}, &schemas.SpaceMember{}); err != nil {
		t.Fatal(err)
	}
	if err := orm.Exec(`CREATE TABLE porte_identities (user_id integer, provider text, subject text)`).Error; err != nil {
		t.Fatal(err)
	}
	owner := schemas.User{Email: "sara@example.com", Name: "Sara"}
	if err := orm.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	agent, err := schemas.EnsureAgentUser(orm, owner.ID, schemas.AgentFromName("claude"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	service := &Service{orm: orm}
	if _, _, err := service.IdentityForUser(ctx, agent.ID); err == nil {
		t.Fatal("a session naming an agent must not authenticate")
	}
	if id, _, err := service.IdentityForUser(ctx, owner.ID); err != nil || id != strconv.FormatInt(owner.ID, 10) {
		t.Fatalf("the owner must still authenticate: %q %v", id, err)
	}

	store := NewUserStore(orm)
	if _, err := store.FindByEmail(ctx, agent.Email); !stderrors.Is(err, porte.ErrNotFound) {
		t.Fatalf("password login must not find an agent, got %v", err)
	}
	claims := porte.Claims{Provider: "https://sso.example.test", Subject: "x", Email: agent.Email, EmailVerified: true}
	if _, err := store.UpsertFromOIDC(ctx, claims); err == nil || !strings.Contains(err.Error(), "agent") {
		t.Fatalf("OIDC must not sign in as an agent, got %v", err)
	}
	if count, err := store.CountUsers(ctx); err != nil || count != 1 {
		t.Fatalf("agents must not count as users: %d %v", count, err)
	}
}
