package plans

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync/atomic"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/forfait"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const AIAccess = "ai.access"

type Gate struct {
	orm    *gorm.DB
	issuer string
	logger *slog.Logger
	store  *forfait.Store
	active atomic.Bool
}

func New(orm *gorm.DB, issuer string, logger *slog.Logger) *Gate {
	store := forfait.NewStore(forfait.Feature{Key: AIAccess, Kind: forfait.Bool, Default: json.RawMessage("false")})
	store.Issuer = issuer
	store.Logger = logger
	return &Gate{orm: orm, issuer: issuer, logger: logger, store: store}
}

func (g *Gate) Load(ctx context.Context) error {
	var rows []schemas.ForfaitState
	if err := g.orm.WithContext(ctx).Where("id = 1").Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	document := []byte(rows[0].Document)
	if err := g.store.Import(document); err != nil {
		g.logger.Error("plans: saved entitlements could not be restored, enforcing defaults", slog.Any("error", err))
		g.store.Apply(directory.Snapshot{EntitlementsRevision: 1})
		g.active.Store(true)
		return nil
	}
	g.refresh(document)
	return nil
}

func (g *Gate) Apply(ctx context.Context, snapshot directory.Snapshot) error {
	g.store.Apply(snapshot)
	document, err := g.store.Export()
	if err != nil {
		return err
	}
	row := schemas.ForfaitState{ID: 1, Document: string(document)}
	err = g.orm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"document", "updated_at"}),
	}).Create(&row).Error
	if err != nil {
		return err
	}
	g.refresh(document)
	return nil
}

func (g *Gate) Active() bool {
	return g != nil && g.active.Load()
}

func (g *Gate) refresh(document []byte) {
	var state struct {
		Revision int64 `json:"entitlements_revision"`
	}
	if json.Unmarshal(document, &state) == nil && state.Revision > 0 {
		g.active.Store(true)
	}
}

func (g *Gate) UpgradeURL() string {
	return forfait.UpgradeURL(g.issuer, "")
}

func (g *Gate) HasAI(ctx context.Context, userID int64) (bool, error) {
	if !g.Active() {
		return true, nil
	}
	var subjects []string
	err := g.orm.WithContext(ctx).Table("porte_identities").Select("subject").
		Where("user_id = ? AND provider = ?", userID, g.issuer).Scan(&subjects).Error
	if err != nil {
		return false, err
	}
	if len(subjects) == 0 {
		subjects = []string{""}
	}
	for _, subject := range subjects {
		if g.store.Enabled(forfait.Scope{Subject: subject}, AIAccess) {
			return true, nil
		}
	}
	return false, nil
}
