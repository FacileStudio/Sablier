package schemas

import "testing"

func TestAvatarPrecedence(t *testing.T) {
	const porte = "https://porte.facile.studio/media/user-avatars/x.png"

	cases := []struct {
		name       string
		user       User
		wantURL    string
		wantOrigin string
	}{
		{"Porte photo wins over an upload", User{OIDCPictureURL: porte, AvatarUploadPath: "avatars/user-3-1.png"}, porte, "oidc"},
		{"upload is the fallback", User{AvatarUploadPath: "avatars/user-3-1.png"}, "/files/avatars/user-3-1.png", "upload"},
		{"only Porte", User{OIDCPictureURL: porte}, porte, "oidc"},
		{"neither, so the client draws initials", User{}, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.user.Avatar(); got != tc.wantURL {
				t.Errorf("Avatar() = %q, want %q", got, tc.wantURL)
			}
			if got := tc.user.AvatarOrigin(); got != tc.wantOrigin {
				t.Errorf("AvatarOrigin() = %q, want %q", got, tc.wantOrigin)
			}
		})
	}
}

// The join in timeentries reads the avatar in SQL rather than loading the row, so the two
// spellings of the same rule have to agree. This is the test that fails when someone edits
// one and forgets the other.
func TestAvatarSelectExprMatchesAvatar(t *testing.T) {
	orm := openTestDatabase(t)
	if err := orm.AutoMigrate(&User{}); err != nil {
		t.Fatalf("prepare schema: %v", err)
	}

	users := []User{
		{Email: "both@example.com", OIDCPictureURL: "https://porte.facile.studio/media/user-avatars/a.png", AvatarUploadPath: "avatars/user-1-1.png"},
		{Email: "upload@example.com", AvatarUploadPath: "avatars/user-2-1.png"},
		{Email: "oidc@example.com", OIDCPictureURL: "https://porte.facile.studio/media/user-avatars/b.png"},
		{Email: "neither@example.com"},
	}
	for i := range users {
		if err := orm.Create(&users[i]).Error; err != nil {
			t.Fatalf("create %s: %v", users[i].Email, err)
		}
	}

	for _, want := range users {
		var got string
		if err := orm.Model(&User{}).
			Select(AvatarSelectExpr).
			Where("users.id = ?", want.ID).
			Scan(&got).Error; err != nil {
			t.Fatalf("select for %s: %v", want.Email, err)
		}
		if got != want.Avatar() {
			t.Errorf("%s: SQL gave %q, Avatar() gave %q", want.Email, got, want.Avatar())
		}
	}
}

// The four rows here are the four found on the production database. Rows 2 and 4 are the
// reason this test exists: they hold an uploaded avatar with avatar_source empty, because
// they predate that column, and a backfill keyed on avatar_source = 'upload' drops their
// picture without a word.
//
// The row carrying both an upload and a Porte photo keeps its file on disk while still
// rendering the Porte photo.
func TestBackfillAvatarUploadPathKeepsPreSourceUploads(t *testing.T) {
	orm := openTestDatabase(t)
	if err := orm.AutoMigrate(&User{}); err != nil {
		t.Fatalf("prepare schema: %v", err)
	}

	rows := []struct {
		email  string
		url    string
		source string
		oidc   string
		want   string
	}{
		{"oidc-copy@example.com", "/files/avatars/oidc-1-178006.png", "oidc", "https://porte.facile.studio/media/user-avatars/a.png", ""},
		{"legacy-upload@example.com", "/files/avatars/user-2-177802.jpg", "", "", "avatars/user-2-177802.jpg"},
		{"upload-and-sso@example.com", "/files/avatars/user-3-178096.jpg", "upload", "https://porte.facile.studio/media/user-avatars/b.jpeg", "avatars/user-3-178096.jpg"},
		{"no-avatar@example.com", "", "", "", ""},
	}
	for _, row := range rows {
		if err := orm.Exec(
			`INSERT INTO users (email, password_hash, avatar_url, avatar_source, oidc_picture_url) VALUES (?, 'hash', ?, ?, ?)`,
			row.email, row.url, row.source, row.oidc).Error; err != nil {
			t.Fatalf("insert %s: %v", row.email, err)
		}
	}

	if err := backfillAvatarUploadPath(orm); err != nil {
		t.Fatalf("backfill: %v", err)
	}

	for _, row := range rows {
		var got User
		if err := orm.Where("email = ?", row.email).First(&got).Error; err != nil {
			t.Fatalf("read %s: %v", row.email, err)
		}
		if got.AvatarUploadPath != row.want {
			t.Errorf("%s: avatar_upload_path = %q, want %q", row.email, got.AvatarUploadPath, row.want)
		}
	}

	var both User
	if err := orm.Where("email = ?", "upload-and-sso@example.com").First(&both).Error; err != nil {
		t.Fatalf("read both: %v", err)
	}
	if both.Avatar() != "https://porte.facile.studio/media/user-avatars/b.jpeg" {
		t.Errorf("SSO photo should win, got %q", both.Avatar())
	}
}
