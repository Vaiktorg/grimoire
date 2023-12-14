package entities

type Address struct {
	Model
	Addr1, Addr2         string
	City, State, Country string
	Zip                  string
}

// Profile TODO: Migrate table changes
type Profile struct {
	Model
	FirstName string
	Initial   string
	LastName  string
	LastName2 string

	PhoneNumber string // Unique

	AddressID string
	Address   *Address `gorm:"foreignKey:AddressID"`
}
