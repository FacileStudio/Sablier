package schemas

import (
	"crypto/rand"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type Space struct {
	ID          string    `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name;not null"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Space) TableName() string { return "spaces" }

func (s *Space) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = newUUID()
	}
	return nil
}

type SpaceMember struct {
	ID       string    `gorm:"column:id;primaryKey"`
	SpaceID  string    `gorm:"column:space_id;not null;uniqueIndex:idx_space_user"`
	UserID   int64     `gorm:"column:user_id;not null;uniqueIndex:idx_space_user"`
	Role     string    `gorm:"column:role;not null;default:'member'"`
	JoinedAt time.Time `gorm:"column:joined_at;autoCreateTime"`
}

func (SpaceMember) TableName() string { return "space_members" }

func (m *SpaceMember) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = newUUID()
	}
	return nil
}
