package schemas

import "time"

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

type Space struct {
	ID          string    `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `gorm:"column:name;not null"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Space) TableName() string { return "spaces" }

type SpaceMember struct {
	ID       string    `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	SpaceID  string    `gorm:"column:space_id;type:uuid;not null;uniqueIndex:idx_space_user"`
	UserID   int64     `gorm:"column:user_id;not null;uniqueIndex:idx_space_user"`
	Role     string    `gorm:"column:role;not null;default:'member'"`
	JoinedAt time.Time `gorm:"column:joined_at;autoCreateTime"`
}

func (SpaceMember) TableName() string { return "space_members" }
