package gwt

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/vaiktorg/grimoire/util"
	"strings"
	"time"
)

type Spice struct {
	Salt   []byte
	Pepper []byte
}

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

func NewMultiCoder[T any](spice Spice) (*MultiCoder[T], error) {
	mc, err := util.NewMultiCoder[*GWT[T]]()
	if err != nil {
		return nil, err
	}

	return &MultiCoder[T]{spice: spice, mc: mc}, nil
}

func (m *MultiCoder[T]) Encode(tok *GWT[T]) (ret Token, err error) {
	data, err := m.mc.Encode(tok, util.EncodeGob)
	if err != nil {
		return
	}

	// ------------------------------------------------------------------------------------------------
	// Gen Signature
	hashSignature, err := GenSignature(data, m.spice)
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
		Signature: base64.URLEncoding.EncodeToString(hashSignature)}, nil
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

	if hashSignature, _ := GenSignature(tknBuff, m.spice); !bytes.Equal(hashSignature, sigBuff) {
		return nil, errors.New(ErrorInvalidToken)
	}

	ret.Token = token
	return ret, nil
}

func GenSignature(tokenBuff []byte, spice Spice) ([]byte, error) {
	//Generate gwt Signature from decoded payload
	hash := hmac.New(sha256.New, spice.Salt)
	_, err := hash.Write(tokenBuff)
	if err != nil {
		return nil, errors.New(ErrorInvalidToken)
	}

	return hash.Sum(spice.Pepper), nil
}

// =======================================================

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

func (gwt *GWT[T]) ValidateGWT(spice Spice) error {
	return gwt.ValidateGWTWithBody(nil, spice)
}

func (gwt *GWT[T]) ValidateGWTHeader(h func(*Header) error, spice Spice) error {
	err := gwt.ValidateGWT(spice)
	if err != nil {
		return err
	}

	return h(&gwt.Header)
}
func (gwt *GWT[T]) ValidateGWTWithBody(bodyValidHandler func(T) error, spice Spice) error {
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

	return gwt.ValidateSignature(gwt.Token, spice)
}
func (gwt *GWT[T]) ValidateSignature(token string, spice Spice) error {
	// tkn := tknParts[0]
	// sig := tknParts[1]
	tknParts := strings.Split(token, ".")
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

	// ====================================================================================================
	// Token Signature

	nSigBuff, err := GenSignature(tokBuff, spice)
	if err != nil {
		return err
	}

	if !bytes.Equal(nSigBuff, sigBuff) {
		return errors.New(ErrorInvalidToken)
	}

	return nil
}
