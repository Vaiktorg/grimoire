package entities

import (
	"github.com/google/uuid"
	"time"
)

type UserActivityLog struct {
	Model

	IdentityId uuid.UUID
	Identity   *Identity `gorm:"foreignKey:IdentityId"`

	LoggedIn  time.Time
	LoggedOut time.Time

	IPAddress string
	Device    string
}
