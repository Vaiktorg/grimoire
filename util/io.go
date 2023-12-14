package util

import (
	"crypto/cipher"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type IEncoder[T comparable] interface {
	Encode(src, dst []byte)
}

// EncodeB64 struct{} -> JSON -> B64
func EncodeB64[T any](src T) ([]byte, error) {
	s, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}

	if ok := json.Valid(s); !ok {
		return nil, err
	}

	//EncodeB64 to B64 UTF8
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(s)))
	base64.StdEncoding.Encode(dst, s)

	// Print
	return dst, nil
}

// DecodeB64 B64 -> JSON -> struct{}
func DecodeB64[T any](src []byte, res T) error {
	dec, err := base64.StdEncoding.DecodeString(string(src))
	if err != nil {
		return err
	}

	// Validate dst is valid SysCert json
	if ok := json.Valid(dec); !ok {
		return errors.New("invalid json")
	}

	// DecodeB64 to JSON UTF8
	return json.Unmarshal(dec, res)
}

// Save struct to file
func Save[T any](filename string, obj T) error {
	barr, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	if ok := json.Valid(barr); !ok {
		return err
	}

	return os.WriteFile(filename, barr, os.ModePerm)
}

func OpenDecode[T any](filename string, ret T) error {
	barr, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return DecodeB64[T](barr, ret)
}

func SaveEncode[T any](filename string, obj T) error {
	src, err := EncodeB64(obj)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(filename), os.ModePerm)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, src, os.ModePerm)
}

/*
Crypto needs a cipher.Block to encrypt/decrypt.

	The Cipher Blocks available are:
	aes.NewCipher, blowfish.NewCipher, tea.NewCipher,
	cast5.NewCipher, twofish.NewCipher, xtea.NewCipher,
	xts.NewCipher, des.NewCipher, rc4.NewCipher,
	chacha20.NewUnauthenticatedCipher,
*/
type Crypto struct {
	key   []byte
	block cipher.Block
}

// NewCrypto creates and returns a new Crypto.
// The key argument should be the AES key.
// Keys of 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
func NewCrypto(key []byte, block cipher.Block) Crypto {
	return Crypto{
		key:   key,
		block: block,
	}
}

func (c *Crypto) DecryptCFB(dst, src []byte) {
	cfb := cipher.NewCFBDecrypter(c.block, make([]byte, c.block.BlockSize()))
	cfb.XORKeyStream(dst, src)
}

func (c *Crypto) EncryptCFB(src, dst []byte) {
	cfb := cipher.NewCFBEncrypter(c.block, make([]byte, c.block.BlockSize()))
	cfb.XORKeyStream(dst, src)
}
