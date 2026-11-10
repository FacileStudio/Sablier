package auth

import (
	"context"
	"slices"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/usercolor"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte"
)

func TestTheKitLeavesOnlyTheAgentAndLegacySubjectColumnsUnwritten(t *testing.T) {
	accounts, _ := accountsOn(t)
	want := []string{"avatar_url", "avatar_source", "avatar_upload_path", "oidc_subject", "oidc_access_token",
		"oidc_refresh_token", "oidc_token_expiry", "profile_synced_at", "color", "password_hash", "owner_user_id",
		"agent_key"}
	got := accounts.Store().Unwritten()
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("unwritten = %v, want %v; fill a new column in fillNewAccount or accept it here", got, want)
	}
}

func nullColumns(t *testing.T, accounts *Accounts, userID int64) []string {
	t.Helper()
	rows, err := accounts.db.Query(`SELECT key FROM jsonb_each((SELECT to_jsonb(u) FROM users u WHERE id = $1))
		WHERE value = 'null'::jsonb ORDER BY key`, userID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			t.Fatal(err)
		}
		out = append(out, key)
	}
	return out
}

func TestAKitSignUpFillsEveryColumnAGormRowFills(t *testing.T) {
	accounts, orm := accountsOn(t)
	ctx := context.Background()
	old := schemas.User{Email: "old@example.com", Name: "Old", Color: "AD9EF0", Locale: "en"}
	if err := orm.Create(&old).Error; err != nil {
		t.Fatal(err)
	}
	id, err := accounts.UpsertFromOIDC(ctx, porte.Claims{Provider: testIssuer, Subject: "new-sub",
		Email: "New@Example.com", EmailVerified: true, Name: "New"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := nullColumns(t, accounts, id), nullColumns(t, accounts, old.ID); !slices.Equal(got, want) {
		t.Fatalf("NULL columns on a kit account = %v, on a GORM account = %v", got, want)
	}
	var created schemas.User
	if err := orm.First(&created, id).Error; err != nil {
		t.Fatalf("a kit account must load through GORM: %v", err)
	}
	if created.Email != "new@example.com" || created.Locale != "en" || created.Kind != schemas.UserKindHuman ||
		created.CreatedAt.IsZero() || created.WorkdayHours != 8 || created.RateType != "daily" {
		t.Fatalf("new account = %+v", created)
	}
	if _, ok := usercolor.Normalize(created.Color); !ok || created.Color == old.Color {
		t.Fatalf("colour = %q, want a palette colour other than %q", created.Color, old.Color)
	}
}

func TestASignInRewritesTheEmailOnlyWhenVerified(t *testing.T) {
	accounts, orm := accountsOn(t)
	ctx := context.Background()
	claims := porte.Claims{Provider: testIssuer, Subject: "s", Email: "a@example.com", EmailVerified: true}
	id, err := accounts.UpsertFromOIDC(ctx, claims)
	if err != nil {
		t.Fatal(err)
	}
	claims.Email, claims.EmailVerified = "b@example.com", false
	if _, err := accounts.UpsertFromOIDC(ctx, claims); err != nil {
		t.Fatal(err)
	}
	var user schemas.User
	if err := orm.First(&user, id).Error; err != nil || user.Email != "a@example.com" {
		t.Fatalf("an unverified address replaced the email: %q %v", user.Email, err)
	}
}
