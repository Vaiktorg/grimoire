package src

import "github.com/vaiktorg/grimoire/authentity/src/entities"

type RegisterRequest struct {
	Account entities.Account
	Profile entities.Profile
}

type LoginRequest struct {
	Username string
	Email    string
	Password string
}
