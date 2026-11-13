package schemas

import (
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var queryModels = []any{
	&User{}, &Project{}, &Task{}, &TimeEntry{}, &AppSetting{}, &PushSubscription{}, &Space{},
	&SpaceMember{}, &AntenneProcessedEvent{}, &Heartbeat{}, &DraftDismissal{}, &ForfaitState{},
}

var timeType = reflect.TypeFor[time.Time]()

func compatibleTypes(field reflect.Type) []string {
	for field.Kind() == reflect.Pointer {
		field = field.Elem()
	}
	if field == timeType {
		return []string{"timestamp with time zone"}
	}
	switch field.Kind() {
	case reflect.String:
		return []string{"text", "character varying", "uuid"}
	case reflect.Bool:
		return []string{"boolean"}
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
		return []string{"bigint"}
	case reflect.Int32, reflect.Uint32:
		return []string{"integer", "bigint"}
	case reflect.Int16, reflect.Uint16:
		return []string{"smallint", "integer", "bigint"}
	case reflect.Float32, reflect.Float64:
		return []string{"numeric", "double precision", "real"}
	default:
		return nil
	}
}

func migratedColumnTypes(t *testing.T, db *gorm.DB, table string) map[string]string {
	t.Helper()
	var rows []struct {
		ColumnName string
		DataType   string
	}
	err := db.Raw(`SELECT column_name, data_type FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = ?`, table).Scan(&rows).Error
	if err != nil {
		t.Fatalf("read the columns of %s: %v", table, err)
	}
	columns := map[string]string{}
	for _, row := range rows {
		columns[row.ColumnName] = row.DataType
	}
	return columns
}

// TestGORMModelsMatchTheMigratedSchema checks every column a GORM model reads
// or writes exists in the schema the migrations build, with a type the field
// can hold. AutoMigrate no longer adds a missing column, so a field added to a
// struct without a migration fails here instead of at query time.
func TestGORMModelsMatchTheMigratedSchema(t *testing.T) {
	db := testdb.Migrated(t)
	for _, model := range queryModels {
		parsed, err := schema.Parse(model, &sync.Map{}, db.NamingStrategy)
		if err != nil {
			t.Fatalf("parse %T: %v", model, err)
		}
		columns := migratedColumnTypes(t, db, parsed.Table)
		if len(columns) == 0 {
			t.Errorf("%T: table %s does not exist", model, parsed.Table)
			continue
		}
		for _, field := range parsed.Fields {
			if field.DBName == "" {
				continue
			}
			got, ok := columns[field.DBName]
			if !ok {
				t.Errorf("%s.%s: no such column", parsed.Table, field.DBName)
				continue
			}
			want := compatibleTypes(field.FieldType)
			if !slices.Contains(want, got) {
				t.Errorf("%s.%s: column is %s, field %s needs one of %v", parsed.Table, field.DBName, got, field.FieldType, want)
			}
		}
	}
}
