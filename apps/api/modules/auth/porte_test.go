package auth

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const issuer = "https://registre.example.test"

func linkingStore(t *testing.T) (*UserStore, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	orm, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := orm.AutoMigrate(&schemas.User{}); err != nil {
		t.Fatal(err)
	}
	if err := orm.Exec(`CREATE TABLE porte_identities (user_id integer, provider text, subject text)`).Error; err != nil {
		t.Fatal(err)
	}
	return NewUserStore(orm), orm
}

func claimsFor(subject, email string, verified bool) porte.Claims {
	return porte.Claims{Provider: issuer, Subject: subject, Email: email, EmailVerified: verified, Locale: "de"}
}

func TestFirstLoginCreatesAccountWithLocale(t *testing.T) {
	store, orm := linkingStore(t)
	id, err := store.UpsertFromOIDC(context.Background(), claimsFor("sub-1", "Ada@Example.test", true))
	if err != nil {
		t.Fatal(err)
	}
	var user schemas.User
	if err := orm.First(&user, id).Error; err != nil {
		t.Fatal(err)
	}
	if user.Email != "ada@example.test" || user.Locale != "de" {
		t.Fatalf("want a lowercased email and the claim locale, got %q %q", user.Email, user.Locale)
	}
}

func TestVerifiedEmailLinksAnUnlinkedAccount(t *testing.T) {
	store, orm := linkingStore(t)
	existing := schemas.User{Email: "ada@example.test", Name: "Ada"}
	if err := orm.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}
	id, err := store.UpsertFromOIDC(context.Background(), claimsFor("sub-1", "ada@example.test", true))
	if err != nil || id != existing.ID {
		t.Fatalf("a verified address must link the existing account: %d %v", id, err)
	}
}

func TestUnverifiedEmailNeverLinks(t *testing.T) {
	store, orm := linkingStore(t)
	if err := orm.Create(&schemas.User{Email: "ada@example.test"}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := store.UpsertFromOIDC(context.Background(), claimsFor("sub-1", "ada@example.test", false)); err == nil {
		t.Fatal("an unverified address must not link an existing account")
	}
}

func TestSecondSubjectOfTheSameIssuerIsRefused(t *testing.T) {
	store, orm := linkingStore(t)
	existing := schemas.User{Email: "ada@example.test"}
	if err := orm.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}
	if err := orm.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, 'sub-owner')`, existing.ID, issuer).Error; err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := store.UpsertFromOIDC(ctx, claimsFor("sub-newcomer", "ada@example.test", true)); err == nil {
		t.Fatal("a recycled address must not be handed to another subject of the same issuer")
	}
	if id, err := store.UpsertFromOIDC(ctx, claimsFor("sub-owner", "ada@example.test", true)); err != nil || id != existing.ID {
		t.Fatalf("the owner must still sign in: %d %v", id, err)
	}
}

func TestVerifiedEmailLinksAMixedCaseStoredAddress(t *testing.T) {
	store, orm := linkingStore(t)
	existing := schemas.User{Email: "Ada@Example.test", Name: "Ada"}
	if err := orm.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}
	id, err := store.UpsertFromOIDC(context.Background(), claimsFor("sub-1", "ada@example.test", true))
	if err != nil || id != existing.ID {
		t.Fatalf("want the existing account %d, got %d, %v", existing.ID, id, err)
	}
}
