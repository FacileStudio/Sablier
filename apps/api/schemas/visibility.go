package schemas

import (
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

// AllSpaces is the space filter value that selects personal rows plus every space the
// caller belongs to.
const AllSpaces = "all"

// ErrSpaceNotVisible is returned when a space filter names a space the caller is not in.
var ErrSpaceNotVisible = errors.New("space is not visible to the user")

// MemberSpaces is the subquery of the space ids userID belongs to. space_members.space_id
// is text while every other space_id column is uuid, so callers compare through
// CAST(... AS TEXT), which both PostgreSQL and SQLite accept.
func MemberSpaces(db *gorm.DB, userID int64) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).
		Model(&SpaceMember{}).Select("space_id").Where("user_id = ?", userID)
}

// Rows names a table's space column and the condition that makes a personal row, one
// with no space, belong to the caller. Personal rows are private: a project is the
// caller's when they own it, an entry when they or one of their agents booked it.
type Rows struct {
	Space    string
	personal string
}

var (
	// Projects scopes the projects table.
	Projects = Rows{Space: "projects.space_id", personal: "projects.owner_id = @user"}
	// Entries scopes the time_entries table.
	Entries = Rows{Space: "time_entries.space_id", personal: "(time_entries.user_id = @user OR time_entries.user_id IN " +
		"(SELECT id FROM users WHERE kind = 'agent' AND owner_user_id = @user))"}
)

// VisibleSpaces narrows query to the caller's personal rows and the rows of every space
// userID belongs to.
func VisibleSpaces(query *gorm.DB, rows Rows, userID int64) *gorm.DB {
	return query.Where("(("+rows.Space+" IS NULL AND "+rows.personal+") OR CAST("+rows.Space+" AS TEXT) IN (@members))",
		sql.Named("user", userID), sql.Named("members", MemberSpaces(query, userID)))
}

// ScopeSpace applies a list's space filter: empty keeps the caller's personal rows,
// AllSpaces keeps every visible row, and any other value keeps that one space after
// checking that userID belongs to it. A space the user is not in returns ErrSpaceNotVisible.
func ScopeSpace(query *gorm.DB, rows Rows, userID int64, spaceID string) (*gorm.DB, error) {
	switch spaceID {
	case "":
		return query.Where(rows.Space+" IS NULL AND "+rows.personal, sql.Named("user", userID)), nil
	case AllSpaces:
		return VisibleSpaces(query, rows, userID), nil
	}
	if err := RequireMember(query, userID, spaceID); err != nil {
		return nil, err
	}
	return query.Where("CAST("+rows.Space+" AS TEXT) = ?", spaceID), nil
}

// RequireMember returns ErrSpaceNotVisible unless userID belongs to spaceID.
func RequireMember(db *gorm.DB, userID int64, spaceID string) error {
	var count int64
	err := db.Session(&gorm.Session{NewDB: true}).Model(&SpaceMember{}).
		Where("space_id = ? AND user_id = ?", spaceID, userID).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrSpaceNotVisible
	}
	return nil
}

// VisibleProject loads projectID when userID can see it, or returns gorm.ErrRecordNotFound.
func VisibleProject(db *gorm.DB, userID, projectID int64) (*Project, error) {
	var project Project
	err := VisibleSpaces(db.Model(&Project{}), Projects, userID).
		Where("projects.id = ?", projectID).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}
