package models

import (
	"github.com/vaiktorg/grimoire/gwt"
)

type Identity struct {
	ID        string         `json:"id"`
	Profile   *Profile       `json:"profile"`
	Account   *Account       `json:"account"`
	Resources *gwt.Resources `json:"roles"`
}
