package schemas

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBackfillFacileIDs(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:facileid?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Project{}, &Task{}); err != nil {
		t.Fatal(err)
	}
	keep := "fac_keep"
	db.Create(&Project{Name: "a"})
	db.Create(&Project{Name: "b", FacileID: &keep})
	db.Create(&Task{ProjectID: 1, Name: "t"})

	if err := backfillFacileIDs(db); err != nil {
		t.Fatal(err)
	}
	var missing int64
	db.Model(&Project{}).Where("facile_id IS NULL").Count(&missing)
	db.Model(&Task{}).Where("facile_id IS NULL").Count(&missing)
	if missing != 0 {
		t.Fatalf("%d rows still without facile_id", missing)
	}
	var b Project
	db.Where("name = ?", "b").First(&b)
	if *b.FacileID != keep {
		t.Fatal("existing facile_id overwritten")
	}
}
