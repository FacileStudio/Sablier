package schemas

import (
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
)

const testIssuer = "https://porte.test/application/o/sablier/"

// A user who has not signed in since Registre began naming a picture for every human
// must still get one, derived from the identity row. Running it twice proves the boot
// backfill is idempotent, and the user without an identity proves it touches nobody else.
func TestBackfillOIDCPicturesDerivesTheRegistreURL(t *testing.T) {
	db := testdb.Migrated(t)
	statements := []string{
		`INSERT INTO users (id, email, name, oidc_picture_url, color, created_at)
		 VALUES (1, 'camille@facile.studio', 'Camille', 'https://porte.test/old-upload.png', '#fff', now()),
		        (2, 'noah@facile.studio', 'Noah', NULL, '#eee', now())`,
		`INSERT INTO porte_identities (user_id, provider, subject) VALUES (1, '` + testIssuer + `', 'sub-1')`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("seed: %v\n%s", err, statement)
		}
	}
	for range 2 {
		if err := backfillOIDCPictures(db, testIssuer); err != nil {
			t.Fatalf("backfill: %v", err)
		}
	}

	want := map[int64]string{1: "https://porte.test/application/o/sablier/avatars/sub-1", 2: ""}
	for id, url := range want {
		var got string
		if err := db.Raw(`SELECT coalesce(oidc_picture_url, '') FROM users WHERE id = ?`, id).Scan(&got).Error; err != nil {
			t.Fatalf("read user %d: %v", id, err)
		}
		if got != url {
			t.Errorf("user %d: oidc_picture_url = %q, want %q", id, got, url)
		}
	}
}
