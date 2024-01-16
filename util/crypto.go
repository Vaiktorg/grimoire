package util

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"github.com/vaiktorg/grimoire/uid"
	"io"
)

var (
	AESKey = uid.NewUID(32)
)

type Crypto struct {
	block cipher.Block
}

func NewCrypto(block cipher.Block) (*Crypto, error) {

	return &Crypto{block: block}, nil
}

func (c *Crypto) EncryptCFB(src []byte) ([]byte, error) {
	iv := make([]byte, c.block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBEncrypter(c.block, iv)
	dst := make([]byte, len(src))
	cfb.XORKeyStream(dst, src)

	return append(iv, dst...), nil
}

func (c *Crypto) DecryptCFB(src []byte) ([]byte, error) {
	if len(src) < c.block.BlockSize() {
		return nil, errors.New("ciphertext too short")
	}

	iv := src[:c.block.BlockSize()]
	src = src[c.block.BlockSize():]

	cfb := cipher.NewCFBDecrypter(c.block, iv)
	dst := make([]byte, len(src))
	cfb.XORKeyStream(dst, src)

	return dst, nil
}
