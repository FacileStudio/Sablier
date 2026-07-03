package schemas

import "time"

type PoolOutbox struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Channel   string    `gorm:"column:channel"`
	Payload   string    `gorm:"column:payload"`
	Attempts  int       `gorm:"column:attempts;not null;default:0"`
	LastError string    `gorm:"column:last_error"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (PoolOutbox) TableName() string { return "pool_outbox" }

type PoolProcessedEvent struct {
	IdempotencyKey string    `gorm:"column:idempotency_key;primaryKey"`
	ProcessedAt    time.Time `gorm:"column:processed_at;autoCreateTime"`
}

func (PoolProcessedEvent) TableName() string { return "pool_processed_events" }
