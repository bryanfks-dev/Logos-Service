package entity

import (
	"time"
	"uuid"

	"gorm.io/gorm"
)

type SyncOutbox struct {
	ID         string    `gorm:"primaryKey;type:varchar(36);column:id"`
	EntityType string    `gorm:"column:entity_type;type:varchar(100);not null;index:idx_sync_outbox_queue"`
	EntityID   string    `gorm:"column:entity_id;type:varchar(36);not null"`
	Operation  string    `gorm:"column:operation;type:varchar(50);not null"`
	QueuedAt   time.Time `gorm:"column:queued_at;not null;index:idx_sync_outbox_queue"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (SyncOutbox) TableName() string {
	return "sync_outboxes"
}

func (e *SyncOutbox) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewV7().String()
	}

	if e.QueuedAt.IsZero() {
		e.QueuedAt = time.Now()
	}

	return nil
}

type SyncOutboxEntry = SyncOutbox
