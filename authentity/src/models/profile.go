package models

type Address struct {
	ID      string `json:"id"`
	Addr1   string `json:"addr1"`
	Addr2   string `json:"addr2"`
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
	Zip     string `json:"zip"`
}

type Profile struct {
	ID string `json:"id"`

	FirstName string `json:"firstName"`
	Initial   string `json:"initial"`
	LastName  string `json:"lastName"`
	LastName2 string `json:"lastName2"`

	PhoneNumber string `json:"phone"` // Unique

	Address *Address `json:"address"`
}
