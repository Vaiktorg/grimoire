package gwt

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/vaiktorg/grimoire/util"
	"strings"
	"time"
)

//go:embed salt.hash
var salt []byte

//go:embed pepper.hash
var pepper []byte

//go:embed key.hash
var HashKey []byte

type Spice struct {
	Salt   []byte
	Pepper []byte
}

var spice = Spice{
	Salt:   salt,
	Pepper: pepper,
}

// ====================================================================================================

type MultiCoder[T any] struct {
	spice Spice
	mc    *util.MultiCoder[*GWT[T]]
}

const (
	ErrorNoTokenFound = "token not found"
	ErrorInvalidToken = "token is invalid"
	ErrorTokenExpired = "token has expired"
)

const TokenExpireTime = time.Minute * 15

func NewMultiCoder[T any]() (*MultiCoder[T], error) {
	mc, err := util.NewMultiCoder[*GWT[T]]()
	if err != nil {
		return nil, err
	}

	return &MultiCoder[T]{
		spice: spice,
		mc:    mc,
	}, nil
}

type GWT[T any] struct {
	Header Header
	Body   T
	Token  string
}
type Header struct {
	Issuer    []byte    // where the token originated
	Recipient []byte    // who the token belongs to
	Expires   time.Time // When it will expirm.
}

// Token gets delivered to the user.
type Token struct {
	// "1a2.b3c.4d5" [data -> byte -> b64].
	// Return this to requester.
	Token string

	// ___.OOO  Last section of Token.
	// Only accessible when decoded from token string
	Signature string
}

func (m *MultiCoder[T]) Encode(tok *GWT[T]) (ret Token, err error) {
	data, err := m.mc.Encode(tok, util.EncodeGob)
	if err != nil {
		return
	}

	// ------------------------------------------------------------------------------------------------
	// Gen Signature [64]byte 128bit
	hashSignature, err := GenSignature(HashKey, data)
	if err != nil {
		return
	}

	// ------------------------------------------------------------------------------------------------
	// Encode to B64
	b64value := base64.URLEncoding.EncodeToString(data)
	b64signature := base64.URLEncoding.EncodeToString(hashSignature)

	// ------------------------------------------------------------------------------------------------
	// Results in token "b64Data.b64Signature"
	return Token{
		Token: strings.Join([]string{
			b64value,
			b64signature,
		}, "."),
		Signature: hex.EncodeToString(hashSignature)}, nil
}
func (m *MultiCoder[T]) Decode(token string) (*GWT[T], error) {
	if token == "" {
		return nil, errors.New(ErrorNoTokenFound)
	}

	tknParts := strings.Split(token, ".")
	if len(tknParts) != 2 {
		return nil, errors.New(ErrorInvalidToken)
	}

	tknBuff, err := base64.URLEncoding.DecodeString(tknParts[0])
	if err != nil {
		return nil, err
	}

	ret, err := m.mc.Decode(tknBuff, util.DecodeGob)
	if err != nil {
		return nil, err
	}

	// Now validate the signature
	sigBuff, err := base64.URLEncoding.DecodeString(tknParts[1])
	if err != nil {
		return nil, err
	}

	if hashSignature, _ := GenSignature(HashKey, tknBuff); !bytes.Equal(hashSignature, sigBuff) {
		return nil, errors.New(ErrorInvalidToken)
	}

	ret.Token = token
	return ret, nil
}

func GenSignature(key []byte, tokenBuff []byte) ([]byte, error) {
	m := hmac.New(sha512.New, key)

	//Generate gwt Signature from decoded payload
	_, err := m.Write(append(tokenBuff, append(spice.Salt[:], spice.Pepper[:]...)...))
	if err != nil {
		return nil, errors.New(ErrorInvalidToken)
	}

	return m.Sum(nil), nil
}

// =======================================================

func ValidateGWT[T any](gwt *GWT[T]) error {
	return ValidateGWTWithBody(gwt, nil)
}
func ValidateGWTHeader[T any](gwt *GWT[T], h func(*Header) error) error {
	err := ValidateGWT[T](gwt)
	if err != nil {
		return err
	}

	return h(&gwt.Header)
}

func ValidateGWTWithBody[T any](gwt *GWT[T], bodyValidHandler func(T) error) error {
	if gwt.Header.Issuer == nil || gwt.Header.Recipient == nil || gwt.Header.Expires.IsZero() {
		return errors.New(ErrorInvalidToken)
	}

	if time.Now().UTC().After(gwt.Header.Expires) {
		return errors.New(ErrorTokenExpired)
	}

	if bodyValidHandler != nil {
		if err := bodyValidHandler(gwt.Body); err != nil {
			return err
		}
	}

	return ValidateSignature[T](gwt)
}
func ValidateSignature[T any](gwt *GWT[T]) error {
	// tkn := tknParts[0]
	// sig := tknParts[1]
	tknParts := strings.Split(gwt.Token, ".")
	if len(tknParts) != 2 {
		return errors.New(ErrorInvalidToken)
	}

	tokBuff, err := base64.URLEncoding.DecodeString(tknParts[0])
	if err != nil {
		return err
	}

	sigBuff, err := base64.URLEncoding.DecodeString(tknParts[1])
	if err != nil {
		return err
	}

	copyGWT := *gwt
	copyGWT.Token = ""

	// ====================================================================================================
	// GWTs own signature

	buff := bytes.NewBuffer([]byte{})
	defer buff.Reset()

	err = json.NewEncoder(buff).Encode(copyGWT)
	if err != nil {
		return err
	}

	//====================================================================================================
	//Token Signature

	nSigBuff, err := GenSignature(HashKey, tokBuff)
	if err != nil {
		return err
	}

	if !hmac.Equal(nSigBuff, sigBuff) {
		return errors.New(ErrorInvalidToken)
	}

	return nil
}
