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
	if err := db.AutoMigrate(&Project{}, &Task{}, &TimeEntry{}); err != nil {
		t.Fatal(err)
	}
	keep := "fac_keep"
	db.Create(&Project{Name: "a"})
	db.Create(&Project{Name: "b", FacileID: &keep})
	db.Create(&Task{ProjectID: 1, Name: "t"})
	db.Create(&TimeEntry{ProjectID: 1, TaskID: 1, UserID: 1})

	if err := BackfillFacileIDs(db); err != nil {
		t.Fatal(err)
	}
	var missing int64
	db.Model(&Project{}).Where("facile_id IS NULL").Count(&missing)
	db.Model(&Task{}).Where("facile_id IS NULL").Count(&missing)
	var entries int64
	db.Model(&TimeEntry{}).Where("facile_id IS NULL").Count(&entries)
	if missing+entries != 0 {
		t.Fatalf("%d rows still without facile_id", missing)
	}
	var b Project
	db.Where("name = ?", "b").First(&b)
	if *b.FacileID != keep {
		t.Fatal("existing facile_id overwritten")
	}
}
