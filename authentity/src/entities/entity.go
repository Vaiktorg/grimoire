package entities

import (
	"github.com/vaiktorg/grimoire/uid"
	"time"

	"gorm.io/gorm"
)

// Entity contains common columns for all tables.
type Entity struct {
	ID        string     `gorm:"type:text;primary_key;unique;" json:"id,omitempty"`
	CreatedAt time.Time  `sql:"index" json:"created_at"`
	UpdatedAt time.Time  `sql:"index" json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate will set a UID rather than numeric ID.
func (u *Entity) BeforeCreate(_ *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = string(uid.New())
	}

	return nil
}
