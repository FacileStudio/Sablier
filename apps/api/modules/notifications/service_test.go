package notifications

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/schemas"

	webpush "github.com/SherClockHolmes/webpush-go"
	"gorm.io/gorm"
)

type reminderRun struct {
	orm  *gorm.DB
	svc  *Service
	logs *bytes.Buffer
	hits *atomic.Int32
}

func newReminderRun(t *testing.T, status int) reminderRun {
	t.Helper()
	orm := testdb.Migrated(t)
	hits := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(status)
		if _, err := w.Write([]byte(`{"errno":106}`)); err != nil {
			t.Errorf("answer the push: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	seedRunningTimer(t, orm, server.URL+"/push/abc")

	private, public, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatalf("generate vapid keys: %v", err)
	}
	logs := &bytes.Buffer{}
	svc := NewService(orm, public, private, "mailto:ops@example.com", slog.New(slog.NewJSONHandler(logs, nil)))
	return reminderRun{orm: orm, svc: svc, logs: logs, hits: hits}
}

func seedRunningTimer(t *testing.T, orm *gorm.DB, endpoint string) {
	t.Helper()
	user := schemas.User{Email: "push@example.com", Locale: "en"}
	if err := orm.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := orm.Create(&schemas.TimeEntry{UserID: user.ID, StartedAt: time.Now().Add(-2 * time.Hour)}).Error; err != nil {
		t.Fatalf("create time entry: %v", err)
	}
	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate subscription key: %v", err)
	}
	if err := orm.Create(&schemas.PushSubscription{
		UserID:   user.ID,
		Endpoint: endpoint,
		P256DH:   base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()),
		Auth:     base64.RawURLEncoding.EncodeToString([]byte(rand.Text())[:16]),
	}).Error; err != nil {
		t.Fatalf("create subscription: %v", err)
	}
}

func (r reminderRun) subscriptions(t *testing.T) int64 {
	t.Helper()
	var count int64
	if err := r.orm.Model(&schemas.PushSubscription{}).Count(&count).Error; err != nil {
		t.Fatalf("count subscriptions: %v", err)
	}
	return count
}

func assertNoNullError(t *testing.T, logs string) {
	t.Helper()
	if strings.Contains(logs, `"error":null`) {
		t.Fatalf("a nil error was logged:\n%s", logs)
	}
}

func TestReminderDeletesASubscriptionThePushServiceReportsGone(t *testing.T) {
	run := newReminderRun(t, http.StatusGone)

	sent, failed := run.svc.SendActiveTimerReminders(context.Background())

	if sent != 0 || failed != 1 || run.hits.Load() != 1 {
		t.Fatalf("sent=%d failed=%d hits=%d, want 0 1 1", sent, failed, run.hits.Load())
	}
	if n := run.subscriptions(t); n != 0 {
		t.Fatalf("%d subscriptions left, want the gone one deleted", n)
	}
	logs := run.logs.String()
	assertNoNullError(t, logs)
	if !strings.Contains(logs, `"status":410`) || strings.Contains(logs, `"level":"ERROR"`) {
		t.Fatalf("want the 410 logged below ERROR:\n%s", logs)
	}
}

func TestReminderKeepsTheSubscriptionAndBacksOffOnAServerError(t *testing.T) {
	run := newReminderRun(t, http.StatusInternalServerError)

	run.svc.SendActiveTimerReminders(context.Background())
	run.svc.SendActiveTimerReminders(context.Background())

	if n := run.subscriptions(t); n != 1 {
		t.Fatalf("%d subscriptions left, want the row kept", n)
	}
	if hits := run.hits.Load(); hits != 1 {
		t.Fatalf("push service hit %d times in a row, want 1 then an hour of backoff", hits)
	}
	logs := run.logs.String()
	assertNoNullError(t, logs)
	if !strings.Contains(logs, `"level":"WARN"`) || !strings.Contains(logs, `"status":500`) || strings.Contains(logs, `"level":"ERROR"`) {
		t.Fatalf("want one WARN carrying status 500 and no ERROR:\n%s", logs)
	}
}
