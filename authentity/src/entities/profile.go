package entities

type Address struct {
	Entity
	Addr1, Addr2         *string
	City, State, Country *string
	Zip                  *string
}

type Profile struct {
	Entity
	FirstName *string
	Initial   *string
	LastName  *string

	PhoneNumber *string // Unique

	AddressID string
	Address   *Address `gorm:"foreignKey:AddressID"`

	ProfilePicture *string
}
