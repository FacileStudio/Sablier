package schemas

import (
	"crypto/rand"
	"encoding/hex"

	"gorm.io/gorm"
)

// NewFacileID returns a random "fac_"-prefixed hex identifier.
func NewFacileID() string {
	b := make([]byte, 10)
	rand.Read(b)
	return "fac_" + hex.EncodeToString(b)
}

func backfillFacileIDs(db *gorm.DB) error {
	for _, table := range []string{"projects", "tasks"} {
		var ids []int64
		if err := db.Table(table).Where("facile_id IS NULL OR facile_id = ''").Pluck("id", &ids).Error; err != nil {
			return err
		}
		for _, id := range ids {
			if err := db.Table(table).Where("id = ?", id).Update("facile_id", NewFacileID()).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
