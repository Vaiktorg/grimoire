package entities

import "github.com/vaiktorg/grimoire/bitflag"

type (
	Groups struct {
		Model
		GroupFlags bitflag.Bit
	}

	Permissions struct {
		Model
		Flags bitflag.Bit
	}
)
