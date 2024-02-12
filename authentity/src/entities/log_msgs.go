package entities

import (
	"github.com/vaiktorg/grimoire/uid"
	"time"
)

type UserActivityLog struct {
	Entity

	IdentityId uid.UID   `gorm:"index"`
	Identity   *Identity `gorm:"foreignKey:IdentityId"`

	LoggedIn  time.Time
	LoggedOut time.Time

	IPAddress string
	Device    string
}
