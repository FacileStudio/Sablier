package schemas

import (
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

// VisibleSpaces narrows query to the rows whose column is NULL (personal) or names a space
// userID belongs to. column is a trusted, qualified column name such as "projects.space_id".
func VisibleSpaces(query *gorm.DB, column string, userID int64) *gorm.DB {
	return query.Where("("+column+" IS NULL OR CAST("+column+" AS TEXT) IN (?))", MemberSpaces(query, userID))
}

// ScopeSpace applies a list's space filter: empty keeps personal rows only, AllSpaces keeps
// every visible row, and any other value keeps that one space after checking that userID
// belongs to it. A space the user is not in returns ErrSpaceNotVisible.
func ScopeSpace(query *gorm.DB, column string, userID int64, spaceID string) (*gorm.DB, error) {
	switch spaceID {
	case "":
		return query.Where(column + " IS NULL"), nil
	case AllSpaces:
		return VisibleSpaces(query, column, userID), nil
	}
	if err := RequireMember(query, userID, spaceID); err != nil {
		return nil, err
	}
	return query.Where("CAST("+column+" AS TEXT) = ?", spaceID), nil
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
	err := VisibleSpaces(db.Model(&Project{}), "projects.space_id", userID).
		Where("projects.id = ?", projectID).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}
