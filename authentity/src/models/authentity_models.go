package models

type RegisterRequest struct {
	Account Account
	Profile Profile
}

type LoginRequest struct {
	Username string
	Email    string
	Password string
}
