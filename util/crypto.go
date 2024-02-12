package util

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"github.com/vaiktorg/grimoire/uid"
	"golang.org/x/crypto/blowfish"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"io"
	"io/ioutil"
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

func (c *Crypto) EncryptBlowfish(src []byte) ([]byte, error) {
	block, err := blowfish.NewCipher(AESKey.Bytes())
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, blowfish.BlockSize+len(src))
	iv := ciphertext[:blowfish.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[blowfish.BlockSize:], src)

	return ciphertext, nil
}

func (c *Crypto) DecryptBlowfish(ciphertext []byte) ([]byte, error) {
	block, err := blowfish.NewCipher(AESKey.Bytes())
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < blowfish.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := ciphertext[:blowfish.BlockSize]
	ciphertext = ciphertext[blowfish.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	return ciphertext, nil
}

func (c *Crypto) EncryptGCM(src []byte) ([]byte, error) {
	nonce := make([]byte, c.block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(c.block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, src, nil)
	return append(nonce, ciphertext...), nil
}

func (c *Crypto) EncryptECB(src []byte) ([]byte, error) {
	if len(src)%c.block.BlockSize() != 0 {
		return nil, errors.New("input not full blocks")
	}

	dst := make([]byte, len(src))
	for i := 0; i < len(src); i += c.block.BlockSize() {
		c.block.Encrypt(dst[i:], src[i:])
	}

	return dst, nil
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

func (c *Crypto) DecryptGCM(src []byte) ([]byte, error) {
	aesgcm, err := cipher.NewGCM(c.block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if len(src) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := src[:nonceSize], src[nonceSize:]
	return aesgcm.Open(nil, nonce, ciphertext, nil)
}

func (c *Crypto) DecryptECB(src []byte) ([]byte, error) {
	if len(src)%c.block.BlockSize() != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	dst := make([]byte, len(src))
	for i := 0; i < len(src); i += c.block.BlockSize() {
		c.block.Decrypt(dst[i:], src[i:])
	}

	return dst, nil
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

func EncryptPGP(src []byte, publicKey []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	armoredWriter, err := armor.Encode(buf, "PGP MESSAGE", nil)
	if err != nil {
		return nil, err
	}

	entitiesList, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(publicKey))
	if err != nil {
		return nil, err
	}

	plainWriter, err := openpgp.Encrypt(armoredWriter, entitiesList, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	_, err = plainWriter.Write(src)
	if err != nil {
		return nil, err
	}

	plainWriter.Close()
	armoredWriter.Close()

	return buf.Bytes(), nil
}

func DecryptPGP(ciphertext []byte, privateKey []byte) ([]byte, error) {
	entitiesList, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(privateKey))
	if err != nil {
		return nil, err
	}

	armoredBlock, err := armor.Decode(bytes.NewReader(ciphertext))
	if err != nil {
		return nil, err
	}

	messageDetails, err := openpgp.ReadMessage(armoredBlock.Body, entitiesList, nil, nil)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(messageDetails.UnverifiedBody)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
