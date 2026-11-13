package schemas

import (
	"context"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
)

func TestMigrateBackfillsMissingUserColors(t *testing.T) {
	orm := testdb.Migrated(t)

	alpha := User{Email: "alpha@example.com", PasswordHash: "hash", Color: "AD9EF0"}
	beta := User{Email: "beta@example.com", PasswordHash: "hash"}
	gamma := User{Email: "gamma@example.com", PasswordHash: "hash"}
	for _, user := range []*User{&alpha, &beta, &gamma} {
		if err := orm.Create(user).Error; err != nil {
			t.Fatalf("create user %s: %v", user.Email, err)
		}
	}

	if err := Migrate(context.Background(), orm, "", nil); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var users []User
	if err := orm.Order("email asc").Find(&users).Error; err != nil {
		t.Fatalf("read users: %v", err)
	}

	expected := map[string]string{
		"alpha@example.com": "AD9EF0",
		"beta@example.com":  "F09ED6",
		"gamma@example.com": "EE7E89",
	}
	for _, user := range users {
		if user.Color != expected[user.Email] {
			t.Fatalf("expected %s to have color %s, got %s", user.Email, expected[user.Email], user.Color)
		}
	}
}

func TestBackfillTimeEntryTasksSkipsUnlinkedEntriesWithoutDescription(t *testing.T) {
	orm := testdb.Migrated(t)
	project := Project{Name: "P"}
	orm.Create(&project)
	orm.Create(&TimeEntry{ProjectID: project.ID, UserID: 1})
	orm.Create(&TimeEntry{ProjectID: 999, UserID: 1, LegacyDescription: "ghost"})
	orm.Create(&TimeEntry{ProjectID: project.ID, UserID: 1, LegacyDescription: "legacy"})

	if err := backfillTimeEntryTasks(orm); err != nil {
		t.Fatal(err)
	}
	var tasks []Task
	orm.Find(&tasks)
	if len(tasks) != 1 || tasks[0].Name != "legacy" {
		t.Fatalf("unexpected tasks: %+v", tasks)
	}
}
