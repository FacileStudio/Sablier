// Package oidcavatar reads the profile claims Authentik returns at login.
//
// It used to download the picture and keep a copy on disk. It no longer does: the URL
// Authentik hands over is public, cacheable and served by Porte, so a local copy was a
// cache of something we are not the source of — one that a container rebuild silently
// emptied while the database went on pointing at it. Storing the URL removes the copy,
// the volume it needed, and the SSRF surface that fetching an attacker-influenced URL
// from inside our own network opened.
package oidcavatar

import "strings"

// Profile holds the OIDC profile claims Authentik returns at login.
type Profile struct {
	Name              string
	PreferredUsername string
	GivenName         string
	FamilyName        string
	Picture           string
}

// DisplayName returns the best available human-readable name from the
// profile claims, falling back from Name to PreferredUsername to the
// combined given/family name, or "" if none are set.
func (p Profile) DisplayName() string {
	if p.Name != "" {
		return p.Name
	}
	if p.PreferredUsername != "" {
		return p.PreferredUsername
	}
	full := strings.TrimSpace(p.GivenName + " " + p.FamilyName)
	if full != "" {
		return full
	}
	return ""
}

// PhotoURL returns the picture claim when it is an https URL, and "" otherwise.
//
// Registre names one stable URL per human, <issuer>/avatars/<sub>, which always serves the
// current picture and answers 404 when there is none, so the client renders initials.
// Authentik used to send `data:image/svg+xml;base64,…` placeholders instead; anything that
// is not https is still reported as no photo, since storing it would never render better
// than the initials.
func PhotoURL(pictureClaim string) string {
	if strings.HasPrefix(strings.ToLower(pictureClaim), "https://") {
		return pictureClaim
	}
	return ""
}
