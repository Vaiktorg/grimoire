package gwt

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Spice struct {
	Salt   []byte
	Pepper []byte
}

type (
	Encoder[T any] struct {
		spice Spice
	}

	// Decoder ...
	Decoder[T any] struct {
		spice Spice
	}

	// Value generates the user's Token
	Value struct {
		Issuer    string    // where the token originated
		Recipient string    // who the token belongs to
		Timestamp time.Time // timeout is enforced
	}

	// Token gets delivered to the user.
	Token struct {
		// "1a2.b3c.4d5" [data -> byte -> b64].
		// Return this to requester.
		Token string

		// ___.___.OOO  Last section of Token.
		// Save this in system.
		Signature []byte
	}
)

const (
	ErrorTokenParts        = "token parts are invalid"
	ErrorNoTokProvided     = "no token provided"
	ErrorFailedToEncodeSig = "failed to encode token signature"
	ErrorSignatureNotMatch = "signatures do not match"
)

func NewEncoder[T any](spice Spice) Encoder[T] {
	return Encoder[T]{
		spice: spice,
	}
}
func (e *Encoder[T]) Encode(values T, res func(token Token) error) error {
	encode := make(chan T)
	resChan := make(chan Token)
	err := make(chan error)

	go e.encodeValue(encode, resChan, err)

	encode <- values
	select {
	case er := <-err:
		return er
	case val := <-resChan:
		return res(val)
	}
}
func (e *Encoder[T]) encodeValue(valChan chan T, resChan chan Token, errChan chan error) {

	val := <-valChan
	// Value JSON string
	valueBuffer := new(bytes.Buffer)
	err := json.NewEncoder(valueBuffer).Encode(val)
	if err != nil {
		errChan <- err
		return
	}

	// ------------------------------------------------------------------------------------------------
	// Gen Signature
	hashSignature, err := genSignature(valueBuffer.Bytes(), e.spice)
	if err != nil {
		errChan <- err
		return
	}

	// ------------------------------------------------------------------------------------------------
	// Encode to B64
	b64value, b64signature := encodeB64(valueBuffer.Bytes(), hashSignature)

	// ------------------------------------------------------------------------------------------------
	// Results in token "b64Header.b64Payload.b64Signature"
	resChan <- Token{
		Token: strings.Join([]string{
			string(b64value),
			string(b64signature),
		}, "."),
		Signature: hashSignature}
}
func encodeB64(valueBuffer, hashSignature []byte) (b64value, b64signature []byte) {
	// ----------------------------------------------------------------------------------------------
	// Signature Encoding
	b64signature = make([]byte, base64.URLEncoding.EncodedLen(len(hashSignature)))
	base64.URLEncoding.Encode(b64signature, hashSignature)

	//----------------------------------------------------------------------------------------------
	// Header Encoding
	b64value = make([]byte, base64.URLEncoding.EncodedLen(len(valueBuffer)))
	base64.URLEncoding.Encode(b64value, valueBuffer)

	return b64value, b64signature
}

func NewDecoder[T any](spice Spice) Decoder[T] {
	return Decoder[T]{spice: spice}
}
func (d *Decoder[T]) Decode(token Token, res func(value T) error) error {
	decode := make(chan Token)
	resChan := make(chan T)
	err := make(chan error)

	go d.decodeToken(decode, resChan, err)

	decode <- token

	select {
	case e := <-err:
		return e
	case val := <-resChan:
		return res(val)
	}
}
func (d *Decoder[T]) decodeToken(decode chan Token, resChan chan T, errChan chan error) {
	tkn := <-decode
	if tkn.Token == "" {
		errChan <- errors.New(ErrorNoTokProvided)
		return
	}

	// tkn := tknB64[0]
	// sig := tknB64[1]
	tknB64 := strings.Split(tkn.Token, ".")
	if len(tknB64) < 2 || len(tknB64) > 2 {
		errChan <- errors.New(ErrorTokenParts)
	}

	tknBuff, err := decodeB64(tknB64[0])
	if err != nil {
		errChan <- err
		return
	}

	sigBuff, err := decodeB64(tknB64[1])
	if err != nil {
		errChan <- err
		return
	}

	// ------------------------------------------------------------------------------------------------
	// Signature
	hashSignature, err := genSignature(tknBuff, d.spice)
	if err != nil {
		errChan <- err
		return
	}

	// ------------------------------------------------------------------------------------------------
	// Validate
	if !bytes.Equal(hashSignature, sigBuff) {
		errChan <- errors.New(ErrorSignatureNotMatch)
		return
	}

	// If signatures match, keep going with decoding information.
	// ------------------------------------------------------------------------------------------------
	// Decode data
	var val T
	err = json.NewDecoder(bytes.NewReader(tknBuff)).Decode(&val)
	if err != nil {
		errChan <- err
		return
	}

	resChan <- val
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

func genSignature(tokenBuff []byte, spice Spice) ([]byte, error) {
	//Generate gwt Signature from decoded payload
	hash := hmac.New(sha256.New, spice.Salt)
	_, err := hash.Write(tokenBuff)
	if err != nil {
		return nil, errors.New(ErrorFailedToEncodeSig)
	}

	return hash.Sum(spice.Pepper), nil
}

func (d *Decoder[T]) ValidateSignature(data T, signature []byte) bool {
	valueBuffer := new(bytes.Buffer)
	err := json.NewEncoder(valueBuffer).Encode(data)
	if err != nil {
		return false
	}

	// ------------------------------------------------------------------------------------------------
	// Gen Signature
	hashSig, err := genSignature(valueBuffer.Bytes(), d.spice)
	if err != nil {
		return false
	}

	if !bytes.Equal(signature, hashSig) {
		return false
	}

	return true
}
