package auth

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte"

	"gorm.io/gorm"
)

const testIssuer = "https://sso.example.test"

func accountsOn(t *testing.T) (*Accounts, *gorm.DB) {
	t.Helper()
	orm := testdb.Migrated(t, testIssuer)
	sqlDB, err := orm.DB()
	if err != nil {
		t.Fatal(err)
	}
	accounts, err := NewAccounts(context.Background(), sqlDB, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	return accounts, orm
}

func TestAgentUserCannotSignIn(t *testing.T) {
	accounts, orm := accountsOn(t)
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
	if _, err := service.Identity(ctx, agent.ID); err == nil {
		t.Fatal("a session naming an agent must not authenticate")
	}
	if identity, err := service.Identity(ctx, owner.ID); err != nil || identity.UserID != owner.ID {
		t.Fatalf("the owner must still authenticate: %+v %v", identity, err)
	}

	byEmail := porte.Claims{Provider: testIssuer, Subject: "x", Email: agent.Email, EmailVerified: true}
	if _, err := accounts.UpsertFromOIDC(ctx, byEmail); err == nil || !strings.Contains(err.Error(), "agent") {
		t.Fatalf("a verified email must not sign in as an agent, got %v", err)
	}
	if err := orm.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, 'agent-sub')`,
		agent.ID, testIssuer+"/").Error; err != nil {
		t.Fatal(err)
	}
	bySubject := porte.Claims{Provider: testIssuer, Subject: "agent-sub", Email: "someone@example.com", EmailVerified: true}
	if _, err := accounts.UpsertFromOIDC(ctx, bySubject); err == nil || !strings.Contains(err.Error(), "agent") {
		t.Fatalf("an identity linked to an agent must not sign in, got %v", err)
	}
	var after schemas.User
	if err := orm.First(&after, agent.ID).Error; err != nil || after.Email != agent.Email || after.Name != agent.Name {
		t.Fatalf("a refused sign-in rewrote the agent row: %+v %v", after, err)
	}
	if id, err := accounts.UpsertFromOIDC(ctx, porte.Claims{Provider: testIssuer, Subject: "sara-sub",
		Email: owner.Email, EmailVerified: true}); err != nil || id != owner.ID {
		t.Fatalf("the owner must still sign in: %d %v", id, err)
	}
}
