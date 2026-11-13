package schemas_test

import (
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

func TestPersonalEntriesStayPrivate(t *testing.T) {
	orm := testdb.Migrated(t)
	owner := int64(1)
	rows := []any{
		&schemas.User{ID: 1, Email: "me@example.test"},
		&schemas.User{ID: 2, Email: "other@example.test"},
		&schemas.User{ID: 3, Email: "bot@example.test", Kind: schemas.UserKindAgent, OwnerUserID: &owner},
	}
	for _, row := range rows {
		if err := orm.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, userID := range []int64{1, 2, 3} {
		if err := orm.Create(&schemas.TimeEntry{UserID: userID, ProjectID: 1, StartedAt: time.Now()}).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, spaceFilter := range []string{"", schemas.AllSpaces} {
		query, err := schemas.ScopeSpace(orm.Model(&schemas.TimeEntry{}), schemas.Entries, 1, spaceFilter)
		if err != nil {
			t.Fatal(err)
		}
		var users []int64
		if err := query.Order("user_id").Pluck("user_id", &users).Error; err != nil {
			t.Fatal(err)
		}
		if len(users) != 2 || users[0] != 1 || users[1] != 3 {
			t.Fatalf("filter %q: want my entries and my agent's, got users %v", spaceFilter, users)
		}
	}
}
