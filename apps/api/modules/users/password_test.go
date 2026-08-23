package users

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/local"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testSchema keeps these tests off `public`, which the schemas package drops
// and recreates for its own migration tests. `go test ./...` runs the two
// packages at once, and two processes reaching for the same schema fail as a
// duplicate key on pg_type rather than as anything that names the collision.
const testSchema = "users_password_test"

// The password path is PostgreSQL or nothing. porte/pg is Postgres-only, the
// credential lives in porte_identities rather than on the user row, and the
// primary key on (provider, subject) is half of what makes "one password per
// account" true — none of which the in-memory SQLite the rest of this package
// uses can express. A green run without a database proves nothing here.
func newPasswordStack(t *testing.T) (*Service, *auth.Service, *gorm.DB) {
	t.Helper()
	raw := os.Getenv("SABLIER_TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("SABLIER_TEST_DATABASE_URL is unset")
	}
	bootstrap, err := gorm.Open(postgres.Open(raw), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := bootstrap.Exec("DROP SCHEMA IF EXISTS " + testSchema + " CASCADE; CREATE SCHEMA " + testSchema).Error; err != nil {
		t.Fatalf("reset the schema: %v", err)
	}
	if sqlDB, err := bootstrap.DB(); err == nil {
		defer sqlDB.Close()
	}

	db, err := gorm.Open(postgres.Open(scopedURL(t, raw)), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open scoped: %v", err)
	}
	if err := schemas.Migrate(db, ""); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql handle: %v", err)
	}
	store := portepg.New(sqlDB)
	accounts := auth.NewUserStore(db)
	sessions, err := session.New(porte.Config{}, session.Deps{Sessions: store.Sessions()})
	if err != nil {
		t.Fatalf("session manager: %v", err)
	}
	passwords, err := local.New(
		local.Config{AllowRegistration: true, MinPasswordLength: 8},
		local.Deps{Users: accounts, Identities: store.Identities(), Sessions: sessions, Count: accounts.CountUsers},
	)
	if err != nil {
		t.Fatalf("local kit: %v", err)
	}
	authService := auth.NewService(db, sessions, passwords, nil)
	return NewService(db, t.TempDir(), authService), authService, db
}

// scopedURL points a connection at testSchema through the DSN rather than a
// SET, because search_path is per connection and gorm hands out a pool.
func scopedURL(t *testing.T, raw string) string {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" {
		t.Fatalf("SABLIER_TEST_DATABASE_URL must be a postgres:// URL, got %q", raw)
	}
	query := parsed.Query()
	query.Set("search_path", testSchema)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func register(t *testing.T, authService *auth.Service, email, password string) int64 {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", nil)
	userID, _, err := authService.Register(request.Context(), recorder, request, email, password)
	if err != nil {
		t.Fatalf("register %s: %v", email, err)
	}
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		t.Fatalf("parse user id %q: %v", userID, err)
	}
	return id
}

func signIn(authService *auth.Service, email, password string) error {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	_, _, err := authService.Login(request.Context(), recorder, request, email, password)
	return err
}

// patchMe drives PATCH /users/me the way the router does, through the writer
// and the request: a password change rotates the session and porte sets the
// replacement cookie itself.
func patchMe(service *Service, userID int64, req *UpdateRequest) (*MeResponse, error) {
	request := httptest.NewRequest(http.MethodPatch, "/api/users/me", nil)
	ctx := authcontext.WithIdentity(request.Context(), authcontext.Identity{UserID: strconv.FormatInt(userID, 10)})
	return service.controller.updateMe(httptest.NewRecorder(), request.WithContext(ctx), req)
}

func localIdentities(t *testing.T, db *gorm.DB) []string {
	t.Helper()
	var subjects []string
	if err := db.Raw(`SELECT subject FROM porte_identities WHERE provider = 'local' ORDER BY subject`).Scan(&subjects).Error; err != nil {
		t.Fatalf("read the password identities: %v", err)
	}
	return subjects
}

func pointer(value string) *string { return &value }

// PATCH /users/me {"password": …} used to reach porte's SetPassword directly,
// so a borrowed session was a permanent account takeover rather than a
// temporary one. OWASP ASVS puts the confirmation at L1 (v4 §2.1.6, v5 §6.2.3).
//
// The status is asserted alongside the credential because a change that is
// refused loudly and applied anyway is the failure worth catching: the old
// password must still sign in and the new one must not.
func TestChangingAPasswordWithoutTheCurrentOneIsRefused(t *testing.T) {
	service, authService, _ := newPasswordStack(t)
	userID := register(t, authService, "camille@facile.studio", "correct horse")

	_, err := patchMe(service, userID, &UpdateRequest{Password: pointer("battery staple")})
	if err == nil {
		t.Fatal("a password was replaced without the current one")
	}
	if status := errors.Status(err); status != http.StatusBadRequest {
		t.Fatalf("expected 400 naming the missing field, got %d: %v", status, err)
	}
	if signIn(authService, "camille@facile.studio", "battery staple") == nil {
		t.Fatal("the refused password was applied anyway")
	}
	if err := signIn(authService, "camille@facile.studio", "correct horse"); err != nil {
		t.Fatalf("the refusal took the existing password with it: %v", err)
	}
}

// A wrong current password is 401 and not 400: the caller sent every field the
// route asks for, and the one that failed is a credential.
func TestChangingAPasswordWithTheWrongCurrentOneIsUnauthorized(t *testing.T) {
	service, authService, _ := newPasswordStack(t)
	userID := register(t, authService, "camille@facile.studio", "correct horse")

	_, err := patchMe(service, userID, &UpdateRequest{
		Password:        pointer("battery staple"),
		CurrentPassword: pointer("not the password"),
	})
	if status := errors.Status(err); status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %v", status, err)
	}
	if err := signIn(authService, "camille@facile.studio", "correct horse"); err != nil {
		t.Fatalf("the existing password stopped working: %v", err)
	}
}

// The confirmed change is the one that has to keep working, and it has to keep
// the caller signed in: porte rotates the session rather than dropping it, and
// the token comes back for the clients holding a bearer instead of a cookie.
func TestChangingAPasswordWithTheCurrentOneRotatesTheSession(t *testing.T) {
	service, authService, db := newPasswordStack(t)
	userID := register(t, authService, "camille@facile.studio", "correct horse")

	response, err := patchMe(service, userID, &UpdateRequest{
		Password:        pointer("battery staple"),
		CurrentPassword: pointer("correct horse"),
	})
	if err != nil {
		t.Fatalf("change the password: %v", err)
	}
	if response.Token == "" {
		t.Fatal("no replacement token came back, so a bearer client is signed out by its own change")
	}
	if err := signIn(authService, "camille@facile.studio", "battery staple"); err != nil {
		t.Fatalf("the new password does not sign in: %v", err)
	}
	if signIn(authService, "camille@facile.studio", "correct horse") == nil {
		t.Fatal("the old password still signs in")
	}
	if subjects := localIdentities(t, db); len(subjects) != 1 {
		t.Fatalf("expected one password identity, got %v", subjects)
	}
}

// Changing an address used to strand the credential: users.email moved and
// porte_identities.subject did not, so the old address kept signing in and the
// new one never did — and a later password change inserted a *second* identity
// on the new address, leaving two working passwords on one account.
//
// Keyed on the account id, none of that is reachable: the address is looked up
// and the credential is read by id.
func TestChangingAnEmailKeepsTheCredentialOnTheAccount(t *testing.T) {
	service, authService, db := newPasswordStack(t)
	userID := register(t, authService, "camille@facile.studio", "correct horse")

	if _, err := patchMe(service, userID, &UpdateRequest{Email: pointer("camille@nouveau.studio")}); err != nil {
		t.Fatalf("change the email: %v", err)
	}
	if err := signIn(authService, "camille@nouveau.studio", "correct horse"); err != nil {
		t.Fatalf("the credential was stranded on the old address: %v", err)
	}
	if signIn(authService, "camille@facile.studio", "correct horse") == nil {
		t.Fatal("the old address still signs in")
	}

	if _, err := patchMe(service, userID, &UpdateRequest{
		Password:        pointer("battery staple"),
		CurrentPassword: pointer("correct horse"),
	}); err != nil {
		t.Fatalf("change the password after the email: %v", err)
	}

	subjects := localIdentities(t, db)
	want := strconv.FormatInt(userID, 10)
	if len(subjects) != 1 || subjects[0] != want {
		t.Fatalf("expected exactly one password identity keyed on %q, got %v", want, subjects)
	}
	if err := signIn(authService, "camille@nouveau.studio", "battery staple"); err != nil {
		t.Fatalf("the new password does not sign in: %v", err)
	}
	if signIn(authService, "camille@nouveau.studio", "correct horse") == nil {
		t.Fatal("the old password survived as a second identity")
	}
}

// A federated account adding its first password has nothing to confirm, so the
// current password is genuinely optional rather than merely usually present.
func TestAFirstPasswordNeedsNoCurrentOne(t *testing.T) {
	service, authService, db := newPasswordStack(t)
	var userID int64
	if err := db.Raw(
		`INSERT INTO users (email, name, color, password_hash, rate, rate_type, workday_hours, created_at)
		 VALUES ('noah@facile.studio', 'Noah', 'AD9EF0', '', 0, 'daily', 8, now()) RETURNING id`).
		Scan(&userID).Error; err != nil {
		t.Fatalf("seed an SSO-only account: %v", err)
	}

	if _, err := patchMe(service, userID, &UpdateRequest{Password: pointer("battery staple")}); err != nil {
		t.Fatalf("set a first password: %v", err)
	}
	if err := signIn(authService, "noah@facile.studio", "battery staple"); err != nil {
		t.Fatalf("the first password does not sign in: %v", err)
	}
	if subjects := localIdentities(t, db); len(subjects) != 1 || subjects[0] != strconv.FormatInt(userID, 10) {
		t.Fatalf("expected one identity keyed on the account id, got %v", subjects)
	}
}
