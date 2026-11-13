package schemas

import "time"

// AntenneProcessedEvent records a pooled event's idempotency key once it has
// been handled.
type AntenneProcessedEvent struct {
	IdempotencyKey string    `gorm:"column:idempotency_key;primaryKey"`
	ProcessedAt    time.Time `gorm:"column:processed_at;autoCreateTime"`
}

func (AntenneProcessedEvent) TableName() string { return "antenne_processed_events" }
