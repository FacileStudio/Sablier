package schemas

import "time"

// User is an account with its profile, credentials and billing settings.
type User struct {
	ID               int64     `gorm:"column:id;primaryKey"`
	Email            string    `gorm:"column:email;uniqueIndex"`
	Name             string    `gorm:"column:name"`
	AvatarURL        string    `gorm:"column:avatar_url"`
	AvatarSource     string    `gorm:"column:avatar_source"`
	AvatarUploadPath string    `gorm:"column:avatar_upload_path"`
	OIDCPictureURL   string    `gorm:"column:oidc_picture_url"`
	OIDCSubject      *string   `gorm:"column:oidc_subject;uniqueIndex"`
	OIDCAccessToken  string    `gorm:"column:oidc_access_token"`
	OIDCRefreshToken string    `gorm:"column:oidc_refresh_token"`
	OIDCTokenExpiry  time.Time `gorm:"column:oidc_token_expiry"`
	ProfileSyncedAt  time.Time `gorm:"column:profile_synced_at"`
	Color            string    `gorm:"column:color"`
	PasswordHash     string    `gorm:"column:password_hash"`
	Rate             float64   `gorm:"column:rate;not null;default:0"`
	RateType         string    `gorm:"column:rate_type;not null;default:'daily'"`
	WorkdayHours     float64   `gorm:"column:workday_hours;not null;default:8"`
	AutoConfirm      bool      `gorm:"column:auto_confirm_drafts;not null;default:true"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	Kind             string    `gorm:"column:kind;not null;default:'human'"`
	OwnerUserID      *int64    `gorm:"column:owner_user_id;index;uniqueIndex:idx_users_owner_agent_key"`
	AgentKey         *string   `gorm:"column:agent_key;uniqueIndex:idx_users_owner_agent_key"`
	AgentAvatarURL   string    `gorm:"column:agent_avatar_url;not null;default:''"`
}

const (
	// UserKindHuman is a person who signs in.
	UserKindHuman = "human"
	// UserKindAgent is an AI agent working for its owner. It never signs in and is never billed.
	UserKindAgent = "agent"
)

// IsAgent reports whether the account is an agent user rather than a person.
func (u User) IsAgent() bool { return u.Kind == UserKindAgent }

func (User) TableName() string { return "users" }

const avatarFilePrefix = "/files/"

// AvatarSelectExpr is Avatar() as SQL, for the joins that read a user's picture without
// loading the row. It has to stay in step with Avatar below — hence both being here,
// one above the other, rather than one in Go and one buried in a Select string.
const AvatarSelectExpr = `CASE WHEN users.kind = 'agent' THEN COALESCE(users.agent_avatar_url, '') ` +
	`ELSE COALESCE(NULLIF(users.oidc_picture_url, ''), ` +
	`NULLIF('/files/' || COALESCE(users.avatar_upload_path, ''), '/files/'), '') END`

// Avatar is the picture to render. It is derived from the two sources rather than stored
// alongside them: the Registre URL always wins, and a legacy upload shows only for an
// account with no Registre identity. Registre answers 404 at that URL for a user with no
// picture, and the client's Avatar falls back to initials. Sablier no longer accepts
// uploads; the ones already on disk are kept and still read here. An agent user shows
// only its own picture.
func (u User) Avatar() string {
	if u.IsAgent() {
		return u.AgentAvatarURL
	}
	if u.OIDCPictureURL != "" {
		return u.OIDCPictureURL
	}
	if u.AvatarUploadPath != "" {
		return avatarFilePrefix + u.AvatarUploadPath
	}
	return ""
}

// AvatarOrigin names where Avatar came from, so the client can say *why* uploading is
// unavailable instead of just greying the button out.
func (u User) AvatarOrigin() string {
	switch {
	case u.IsAgent():
		return UserKindAgent
	case u.OIDCPictureURL != "":
		return "oidc"
	case u.AvatarUploadPath != "":
		return "upload"
	default:
		return ""
	}
}
