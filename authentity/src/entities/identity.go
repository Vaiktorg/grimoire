package entities

type Identity struct {
	Entity

	ProfileID string   // Unique
	Profile   *Profile `gorm:"foreignKey:ProfileID"`

	AccountID string
	Account   *Account `gorm:"foreignKey:AccountID"`

	//Devices *Device

	Resources *string
}
