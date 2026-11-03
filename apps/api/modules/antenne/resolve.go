package antenne

import (
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"gorm.io/gorm"
)

// placement is where a project lives locally: a space, or nil for the owner's personal space.
type placement struct {
	spaceID *string
	ownerID int64
}

func (p placement) same(project *schemas.Project) bool {
	if p.spaceID == nil || project.SpaceID == nil {
		return p.spaceID == nil && project.SpaceID == nil && p.ownerID == project.OwnerID
	}
	return *p.spaceID == *project.SpaceID
}

// resolveUser finds a human by porte subject under the issuer, then by case-insensitive email.
func (s *Service) resolveUser(tx *gorm.DB, subject, email string) *int64 {
	if subject != "" && s.issuer != "" {
		var ids []int64
		err := tx.Table("porte_identities").Where("provider = ? AND subject = ?", s.issuer, subject).
			Limit(1).Pluck("user_id", &ids).Error
		if err == nil && len(ids) > 0 {
			return &ids[0]
		}
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil
	}
	var user schemas.User
	if err := tx.Select("id").Where("lower(email) = ? AND kind <> ?", email, schemas.UserKindAgent).
		First(&user).Error; err != nil {
		return nil
	}
	return &user.ID
}

// resolvePlacement maps the wire space onto a local space and owner, per the contract:
// a Registre space id wins, else a personal project lands in its owner's personal space.
func (s *Service) resolvePlacement(tx *gorm.DB, space *spacePayload) (placement, bool) {
	if space == nil {
		return placement{}, false
	}
	owner := s.resolveUser(tx, space.OwnerSubject, space.OwnerEmail)
	if space.RegistreSpaceID == "" {
		if !space.Personal || owner == nil {
			return placement{}, false
		}
		return placement{ownerID: *owner}, true
	}
	var local schemas.Space
	if err := tx.Where("registre_space_id = ?", space.RegistreSpaceID).First(&local).Error; err != nil {
		return placement{}, false
	}
	if owner == nil {
		owner = spaceOwner(tx, local.ID)
	}
	if owner == nil {
		return placement{}, false
	}
	return placement{spaceID: &local.ID, ownerID: *owner}, true
}

func spaceOwner(tx *gorm.DB, spaceID string) *int64 {
	var member schemas.SpaceMember
	err := tx.Where("space_id = ?", spaceID).
		Order("CASE WHEN role = 'owner' THEN 0 ELSE 1 END, joined_at").First(&member).Error
	if err != nil {
		return nil
	}
	return &member.UserID
}

// identityOf returns a user's lowercased email and porte subject under the issuer.
func (s *Service) identityOf(userID int64) (string, string) {
	var user schemas.User
	email := ""
	if s.orm.Select("email").Where("id = ?", userID).First(&user).Error == nil {
		email = strings.ToLower(user.Email)
	}
	if s.issuer == "" {
		return email, ""
	}
	var subjects []string
	s.orm.Table("porte_identities").Where("provider = ? AND user_id = ?", s.issuer, userID).
		Limit(1).Pluck("subject", &subjects)
	if len(subjects) == 0 {
		return email, ""
	}
	return email, subjects[0]
}

// spaceOf describes where a project lives on the wire. It reports false for a project in
// a space Registre does not know, which is never sent.
func (s *Service) spaceOf(project *schemas.Project) (spacePayload, bool) {
	out := spacePayload{}
	out.OwnerEmail, out.OwnerSubject = s.identityOf(project.OwnerID)
	if project.SpaceID == nil {
		out.Personal = true
		return out, true
	}
	var space schemas.Space
	err := s.orm.Where("id = ?", *project.SpaceID).First(&space).Error
	if err != nil || space.RegistreSpaceID == nil || *space.RegistreSpaceID == "" {
		return out, false
	}
	out.RegistreSpaceID = *space.RegistreSpaceID
	return out, true
}
