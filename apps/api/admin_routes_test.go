package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/middleware"
	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/modules/antenne"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/Sablier/apps/api/modules/settings"
	"github.com/FacileStudio/Sablier/apps/api/modules/users"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

const (
	storedWebhookSecret = "stored-webhook-secret"
	storedBusSecret     = "stored-bus-secret"
)

type adminFixture struct {
	db     *gorm.DB
	router chi.Router
	admin  string
	member string
}

func newAdminFixture(t *testing.T) adminFixture {
	t.Helper()
	db := testdb.Migrated(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sessions, err := session.New(porte.Config{}, session.Deps{Sessions: portepg.New(sqlDB).Sessions(), Logger: logger})
	if err != nil {
		t.Fatal(err)
	}
	authService := auth.NewService(db, sessions, logger)
	admins := middleware.Admins{"Admin@Example.com"}
	router := chi.NewRouter()
	users.RegisterRoutes(router, users.NewService(db, authService, admins), authService)
	settings.RegisterRoutes(router, settings.NewService(db), authService, admins)
	antenne.RegisterRoutes(router, antenne.NewService(db, logger), authService, admins)

	stored := schemas.AppSetting{ID: 1, WebhookURL: "https://hooks.example.test", WebhookSecretHeader: "x-sig",
		WebhookSecretValue: storedWebhookSecret, AntenneURL: "https://bus.example.test", AntenneSecret: storedBusSecret}
	if err := db.Create(&stored).Error; err != nil {
		t.Fatal(err)
	}
	return adminFixture{db: db, router: router,
		admin: bearerFor(t, db, authService, "admin@example.com"), member: bearerFor(t, db, authService, "member@example.com")}
}

func bearerFor(t *testing.T, db *gorm.DB, authService *auth.Service, email string) string {
	t.Helper()
	user := schemas.User{Email: email, Name: email}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	token, _, err := authService.Issue(context.Background(), user.ID, "test")
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func (f adminFixture) call(t *testing.T, token, method, path, body string) (int, string) {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	f.router.ServeHTTP(recorder, request)
	return recorder.Code, recorder.Body.String()
}

var adminRoutes = []struct{ method, path, body string }{
	{http.MethodGet, "/settings/", ""},
	{http.MethodPut, "/settings/", `{"webhook_url":"https://hooks.example.test","webhook_secret_header":"x-sig"}`},
	{http.MethodGet, "/antenne/", ""},
	{http.MethodPut, "/antenne/", `{"antenne_url":"https://bus.example.test","antenne_enabled":false}`},
	{http.MethodGet, "/antenne/events", ""},
	{http.MethodPut, "/antenne/events", `{"events":[{"event":"project.created","enabled":true}]}`},
	{http.MethodPost, "/antenne/sync", ""},
}

// Any Registre account can sign in, so the app-wide webhook and bus settings
// must refuse every account that is not on SABLIER_ADMINS.
func TestAppWideSettingsRefuseANonAdmin(t *testing.T) {
	f := newAdminFixture(t)
	for _, route := range adminRoutes {
		status, body := f.call(t, f.member, route.method, route.path, route.body)
		if status != http.StatusForbidden {
			t.Errorf("%s %s as a non-admin = %d, want 403", route.method, route.path, status)
		}
		if strings.Contains(body, storedWebhookSecret) || strings.Contains(body, storedBusSecret) {
			t.Errorf("%s %s leaked a secret to a non-admin", route.method, route.path)
		}
	}
}

// An admin still reads and saves the settings page, and no answer carries a secret.
func TestAppWideSettingsServeAnAdminWithoutSecrets(t *testing.T) {
	f := newAdminFixture(t)
	for _, route := range adminRoutes {
		status, body := f.call(t, f.admin, route.method, route.path, route.body)
		if status == http.StatusForbidden || status == http.StatusUnauthorized {
			t.Errorf("%s %s as an admin = %d", route.method, route.path, status)
		}
		if strings.Contains(body, storedWebhookSecret) || strings.Contains(body, storedBusSecret) {
			t.Errorf("%s %s returned a secret value", route.method, route.path)
		}
	}

	_, body := f.call(t, f.admin, http.MethodGet, "/settings/", "")
	var webhook settings.SettingsResponse
	if err := json.Unmarshal([]byte(body), &webhook); err != nil || !webhook.Settings.WebhookSecretSet || webhook.Settings.WebhookURL == "" {
		t.Fatalf("GET /settings/ = %s, want the URL and webhook_secret_set", body)
	}
	_, body = f.call(t, f.admin, http.MethodGet, "/antenne/", "")
	var pool antenne.PoolSettingsResponse
	if err := json.Unmarshal([]byte(body), &pool); err != nil || !pool.Settings.SecretSet || pool.Settings.URL == "" {
		t.Fatalf("GET /antenne/ = %s, want the URL and antenne_secret_set", body)
	}

	var stored schemas.AppSetting
	if err := f.db.First(&stored, 1).Error; err != nil {
		t.Fatal(err)
	}
	if stored.WebhookSecretValue != storedWebhookSecret || stored.AntenneSecret != storedBusSecret {
		t.Fatal("a save without a secret field must keep the stored secrets")
	}
}

// The client shows the integrations tab from is_admin on /users/me.
func TestMeReportsWhetherTheAccountIsAnAdmin(t *testing.T) {
	f := newAdminFixture(t)
	for token, want := range map[string]bool{f.admin: true, f.member: false} {
		_, body := f.call(t, token, http.MethodGet, "/users/me", "")
		var me users.MeResponse
		if err := json.Unmarshal([]byte(body), &me); err != nil {
			t.Fatalf("GET /users/me = %s: %v", body, err)
		}
		if me.User.IsAdmin != want {
			t.Errorf("%s: is_admin = %v, want %v", me.User.Email, me.User.IsAdmin, want)
		}
	}
}

// Two accounts that share no space see nothing of each other.
func TestUsersRoutesHideAccountsOutsideTheCallersSpaces(t *testing.T) {
	f := newAdminFixture(t)
	_, body := f.call(t, f.member, http.MethodGet, "/users/", "")
	if strings.Contains(body, "admin@example.com") {
		t.Fatalf("GET /users/ as a stranger listed the admin: %s", body)
	}
	var admin schemas.User
	if err := f.db.Where("email = ?", "admin@example.com").First(&admin).Error; err != nil {
		t.Fatal(err)
	}
	status, _ := f.call(t, f.member, http.MethodGet, "/users/"+strconv.FormatInt(admin.ID, 10), "")
	if status != http.StatusNotFound {
		t.Fatalf("GET /users/{id} of a stranger = %d, want 404", status)
	}
}

// The unauthenticated test broadcast pushed any text to every subscriber. It is gone.
func TestTheTestBroadcastIsNotMounted(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/notifications/test-broadcast", strings.NewReader(`{"title":"x"}`))
	testRouter(t).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound && recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /api/notifications/test-broadcast = %d, want 404", recorder.Code)
	}
}
