package schemas

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type myceliumEntry struct {
	ID          int64  `gorm:"column:id"`
	UserID      int64  `gorm:"column:user_id"`
	Description string `gorm:"column:description"`
}

// backfillAgentEntries moves the Mycelium session entries booked to a human before agent
// users existed onto that human's agent user. It is idempotent: a moved entry belongs to
// an agent, which the query skips.
func backfillAgentEntries(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var rows []myceliumEntry
		err := tx.Table("time_entries").
			Select("time_entries.id, time_entries.user_id, time_entries.description").
			Joins("JOIN users ON users.id = time_entries.user_id").
			Where("time_entries.description LIKE ? AND time_entries.facile_id IS NOT NULL AND users.kind <> ?",
				MyceliumDescriptionPrefix+"%", UserKindAgent).
			Order("time_entries.id asc").
			Find(&rows).Error
		if err != nil {
			return err
		}
		agents := map[string]int64{}
		for _, row := range rows {
			if err := moveToAgent(tx, row, agents); err != nil {
				return err
			}
		}
		return nil
	})
}

func moveToAgent(tx *gorm.DB, row myceliumEntry, agents map[string]int64) error {
	identity := AgentFromName(myceliumAgentName(row.Description))
	cacheKey := fmt.Sprintf("%d/%s", row.UserID, identity.Key)
	agentID, ok := agents[cacheKey]
	if !ok {
		agent, err := EnsureAgentUser(tx, row.UserID, identity)
		if err != nil {
			return err
		}
		agentID = agent.ID
		agents[cacheKey] = agentID
	}
	return tx.Table("time_entries").Where("id = ?", row.ID).Update("user_id", agentID).Error
}

func myceliumAgentName(description string) string {
	rest := strings.TrimPrefix(description, MyceliumDescriptionPrefix)
	if end := strings.IndexAny(rest, "@ "); end >= 0 {
		return rest[:end]
	}
	return rest
}
