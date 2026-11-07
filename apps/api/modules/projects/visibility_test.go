package projects

import (
	"context"
	"net/http"
	"slices"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

func TestListProjectsChecksSpaceMembership(t *testing.T) {
	orm := testdb.Open(t)
	service := NewService(orm)
	mine, foreign := schemas.Space{Name: "Mine"}, schemas.Space{Name: "Foreign"}
	for _, value := range []any{&schemas.User{ID: 1, Email: "a@example.test"}, &mine, &foreign} {
		if err := orm.Create(value).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := orm.Create(&schemas.SpaceMember{SpaceID: mine.ID, UserID: 1}).Error; err != nil {
		t.Fatal(err)
	}
	for _, project := range []schemas.Project{
		{Name: "Personal", OwnerID: 1}, {Name: "Team", OwnerID: 1, SpaceID: &mine.ID},
		{Name: "Secret", OwnerID: 2, SpaceID: &foreign.ID},
	} {
		if err := orm.Create(&project).Error; err != nil {
			t.Fatal(err)
		}
	}

	cases := []struct {
		space string
		want  []string
	}{{"", []string{"Personal"}}, {mine.ID, []string{"Team"}}, {schemas.AllSpaces, []string{"Personal", "Team"}}}
	for _, tc := range cases {
		records, err := service.listProjects(context.Background(), 1, tc.space)
		if err != nil {
			t.Fatalf("space %q: %v", tc.space, err)
		}
		var names []string
		for _, record := range records {
			names = append(names, record.Name)
		}
		slices.Sort(names)
		if !slices.Equal(names, tc.want) {
			t.Fatalf("space %q: got %v, want %v", tc.space, names, tc.want)
		}
	}

	_, err := service.listProjects(context.Background(), 1, foreign.ID)
	if errors.Status(err) != http.StatusNotFound {
		t.Fatalf("a space the caller is not in must be a 404, got %v", err)
	}
	_, err = service.createProject(context.Background(), "1", "Sneaky", "", nil, &foreign.ID)
	if errors.Status(err) != http.StatusNotFound {
		t.Fatalf("creating a project in a foreign space must be a 404, got %v", err)
	}
}
