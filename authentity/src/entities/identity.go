package entities

import (
	"github.com/google/uuid"
	"github.com/vaiktorg/grimoire/bitflag"
)

type Identity struct {
	Model

	ProfileID uuid.UUID // Unique
	Profile   *Profile  `gorm:"foreignKey:ProfileID"`

	AccountID string
	Account   *Account `gorm:"foreignKey:AccountID"`

	GroupsID string
	Groups   bitflag.Bit `gorm:"foreignKey:GroupsID"`

	PermissionsID string
	Permissions   bitflag.Bit `gorm:"foreignKey:PermissionsID"`

	Signature []byte
}
