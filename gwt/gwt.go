package gwt

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"
)

type Spice struct {
	Salt   []byte
	Pepper []byte
}

type (
	Encoder[T comparable] struct {
		mu    sync.Mutex
		spice Spice
		buff  *bytes.Buffer
	}
	Decoder[T comparable] struct {
		mu    sync.Mutex
		spice Spice
		buff  *bytes.Buffer
	}
)

const (
	ErrorMalformedToken        = "token parts are invalid"
	ErrorInvalidToken          = "no token provided"
	ErrorInvalidTokenSignature = "invalid token signature"
	ErrorFailedToEncodeSig     = "failed to encode token signature"
	ErrorSignatureNotMatch     = "signatures do not match"
)

const TokenExpireTime = time.Hour * 24

func NewEncoder[T comparable](spice Spice) Encoder[T] {
	return Encoder[T]{
		spice: spice,
		buff:  bytes.NewBuffer([]byte{}),
	}
}
func (e *Encoder[T]) Encode(data T) (ret Token, err error) {
	e.mu.Lock()

	// Header
	err = json.NewEncoder(e.buff).Encode(data)
	if err != nil {
		return
	}

	// ------------------------------------------------------------------------------------------------
	// Gen Signature
	hashSignature, err := GenSignature(e.buff.Bytes(), e.spice)
	if err != nil {
		return
	}

	// ------------------------------------------------------------------------------------------------
	// Encode to B64
	b64value := encodeB64(e.buff.Bytes())
	b64signature := encodeB64(hashSignature)

	// ------------------------------------------------------------------------------------------------
	// Results in token "b64Header.b64Payload.b64Signature"

	e.buff.Reset()
	e.mu.Unlock()
	return Token{
		Token: strings.Join([]string{
			string(b64value),
			string(b64signature),
		}, "."),
		Signature: hashSignature}, nil
}
func encodeB64(buffer []byte) (b64 []byte) {
	b64 = make([]byte, base64.URLEncoding.EncodedLen(len(buffer)))
	base64.URLEncoding.Encode(b64, buffer)

	return
}

func NewDecoder[T comparable](spice Spice) Decoder[T] {
	return Decoder[T]{
		spice: spice,
		buff:  bytes.NewBuffer([]byte{}),
	}
}
func (d *Decoder[T]) Decode(token string) (ret GWT[T], err error) {
	if token == "" {
		return ret, errors.New(ErrorInvalidToken)
	}

	// tkn := tknParts[0]
	// sig := tknParts[1]
	tknParts := strings.Split(token, ".")
	if len(tknParts) != 2 {
		return ret, errors.New(ErrorMalformedToken)
	}

	tknBuff, err := decodeB64(tknParts[0])
	if err != nil {
		return
	}

	sigBuff, err := decodeB64(tknParts[1])
	if err != nil {
		return
	}

	// ------------------------------------------------------------------------------------------------
	// Signature
	hashSignature, err := GenSignature(tknBuff, d.spice)
	if err != nil {
		return
	}

	// ------------------------------------------------------------------------------------------------
	// Validate
	if !bytes.Equal(hashSignature, sigBuff) {
		return ret, errors.New(ErrorSignatureNotMatch)
	}

	// If signatures match, keep going with decoding information.
	// ------------------------------------------------------------------------------------------------
	// Decode data
	err = json.NewDecoder(bytes.NewReader(tknBuff)).Decode(&ret)
	if err != nil {
		return
	}

	ret.Token.Signature = sigBuff
	ret.Token.Token = token

	return ret, nil
}
func decodeB64(gwt string) ([]byte, error) {
	// ------------------------------------------------------------------------------------------------
	// B64 Decoding
	// Signature
	tokenBuff, err := base64.URLEncoding.DecodeString(gwt)
	if err != nil {
		return nil, err
	}

	return tokenBuff, nil
}

func GenSignature(tokenBuff []byte, spice Spice) ([]byte, error) {
	//Generate gwt Signature from decoded payload
	hash := hmac.New(sha256.New, spice.Salt)
	_, err := hash.Write(tokenBuff)
	if err != nil {
		return nil, errors.New(ErrorFailedToEncodeSig)
	}

	return hash.Sum(spice.Pepper), nil
}

// =======================================================

type GWT[T comparable] struct {
	Header Header
	Body   T
	Token  Token
}
type Header struct {
	Issuer    string    // where the token originated
	Recipient string    // who the token belongs to
	Expires   time.Time // When it will expire.
}

// Token gets delivered to the user.
type Token struct {
	// "1a2.b3c.4d5" [data -> byte -> b64].
	// Return this to requester.
	Token string

	// ___.OOO  Last section of Token.
	// Only accessible when decoded from token string
	Signature []byte
}

func (gwt *GWT[T]) ValidateGWT(validationHandler func(T)) error {
	if gwt.Header.Issuer == "" {
		return errors.New("no issuer assigned")
	}

	if gwt.Header.Recipient == "" {
		return errors.New("no recipient assigned")
	}

	// may customize the time validation to more sophisticated rule if needed
	if gwt.Header.Expires.IsZero() {
		return errors.New("no expire date assigned")
	}

	if time.Since(gwt.Header.Expires) >= 15*time.Minute {
		return errors.New("no expire date assigned")
	}

	if validationHandler != nil {
		validationHandler(gwt.Body)
	}

	if gwt.Token.Token == "" {
		return errors.New("no token found")
	}

	if gwt.Token.Signature == nil || len(gwt.Token.Signature) == 0 {
		return errors.New("so signature found")
	}

	return nil
}
