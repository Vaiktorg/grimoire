package entities

type Account struct {
	Model
	Username string // Unique
	Email    string // Unique
	Password string
}
