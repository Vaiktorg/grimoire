package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model contains common columns for all tables.
type Model struct {
	ID        string     `gorm:"type:text;primary_key;unique;" json:"id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (u *Model) BeforeCreate(_ *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return nil
}
