package schemas

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"github.com/FacileStudio/Sablier/apps/api/internal/usercolor"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MyceliumDescriptionPrefix starts the description of every time entry booked from a
// Mycelium agent session, which is what the agent backfill keys on.
const MyceliumDescriptionPrefix = "mycelium: "

const agentKeyMaxLength = 40

// AgentIdentity names one agent of an owner: the key that makes it unique for that
// owner, the display name and the picture.
type AgentIdentity struct {
	Key       string
	Name      string
	AvatarURL string
}

// AgentFromName derives a full identity from a bare agent name such as Mycelium's
// "claude": a slug key, a title-cased display name and a blobatar picture.
func AgentFromName(raw string) AgentIdentity {
	name := AgentDisplayName(raw)
	return AgentIdentity{Key: AgentSlug(raw), Name: name, AvatarURL: BlobAvatarURL(name)}
}

// AgentSlug lowercases raw and collapses every run of other characters into one dash,
// capped at 40 characters. An empty result falls back to "agent".
func AgentSlug(raw string) string {
	slug := strings.Join(agentWords(strings.ToLower(raw)), "-")
	if len(slug) > agentKeyMaxLength {
		slug = strings.TrimRight(slug[:agentKeyMaxLength], "-")
	}
	if slug == "" {
		return "agent"
	}
	return slug
}

// AgentDisplayName title-cases each word of raw, so "claude-code" reads "Claude Code".
func AgentDisplayName(raw string) string {
	words := agentWords(raw)
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	if len(words) == 0 {
		return "Agent"
	}
	return strings.Join(words, " ")
}

// BlobAvatarURL is the blobatar picture the suite gives an agent with no picture of its own.
func BlobAvatarURL(name string) string {
	return "https://blobatar.dev/avatar/" + url.PathEscape(name) + "?gen=2"
}

func agentWords(raw string) []string {
	return strings.FieldsFunc(raw, func(r rune) bool {
		return r > unicode.MaxASCII || !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
}

// EnsureAgentUser returns the owner's agent user for identity.Key, creating it on first
// use. An existing agent takes the latest name and picture. A new agent joins every space
// its owner belongs to at that moment, and only those, so team views show it next to its
// owner. Agent users carry rate 0, no password, no OIDC identity and a synthetic address
// under the reserved .invalid domain, so nothing can sign in as one.
func EnsureAgentUser(tx *gorm.DB, ownerID int64, identity AgentIdentity) (*User, error) {
	agent, err := findAgent(tx, ownerID, identity.Key)
	if err == nil {
		return agent, refreshAgent(tx, agent, identity)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	color, err := usercolor.NextAvailable(context.Background(), tx)
	if err != nil {
		return nil, err
	}
	key := identity.Key
	created := User{
		Email:          fmt.Sprintf("agent-%d-%s@agents.invalid", ownerID, key),
		Name:           identity.Name,
		Color:          color,
		Kind:           UserKindAgent,
		OwnerUserID:    &ownerID,
		AgentKey:       &key,
		AgentAvatarURL: identity.AvatarURL,
	}
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&created)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return findAgent(tx, ownerID, key)
	}
	return &created, joinOwnerSpaces(tx, ownerID, created.ID)
}

func findAgent(tx *gorm.DB, ownerID int64, key string) (*User, error) {
	var agent User
	err := tx.Where("owner_user_id = ? AND agent_key = ? AND kind = ?", ownerID, key, UserKindAgent).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func refreshAgent(tx *gorm.DB, agent *User, identity AgentIdentity) error {
	if agent.Name == identity.Name && agent.AgentAvatarURL == identity.AvatarURL {
		return nil
	}
	agent.Name = identity.Name
	agent.AgentAvatarURL = identity.AvatarURL
	return tx.Model(&User{}).Where("id = ?", agent.ID).
		Updates(map[string]any{"name": identity.Name, "agent_avatar_url": identity.AvatarURL}).Error
}

func joinOwnerSpaces(tx *gorm.DB, ownerID, agentID int64) error {
	var spaceIDs []string
	if err := tx.Model(&SpaceMember{}).Where("user_id = ?", ownerID).Pluck("space_id", &spaceIDs).Error; err != nil {
		return err
	}
	for _, spaceID := range spaceIDs {
		member := SpaceMember{SpaceID: spaceID, UserID: agentID, Role: RoleMember}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&member).Error; err != nil {
			return err
		}
	}
	return nil
}
