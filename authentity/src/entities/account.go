package entities

type Account struct {
	Entity
	Username  string  // Unique
	Email     *string // Unique
	Password  *string
	Signature *string
}
