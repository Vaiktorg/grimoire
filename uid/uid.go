package uid

// Package uid generates a ServiceURL safe string.
//
//  id := Gen(10)
//  fmt.Println(id)
//  // 9BZ1sApAX4

import (
	cr "crypto/rand"
	"encoding/hex"
	"math/rand"
	"time"
)

// http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
const AlphaNumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const UPPERCASEAlphaNumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const Alphabetic = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const Numeric = "0123456789"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var (
	src = rand.NewSource(time.Now().UnixNano())
)

func init() {
	src.Seed(time.Now().UnixNano())
}

type UID string

func New() UID {
	return newBytes(16, AlphaNumeric)
}

// NewUID takes constant letterBytes and returns random string of length n.
func NewUID(n int) UID {
	return newBytes(n, AlphaNumeric)
}
func NewSecure512() (UID, error) {
	// Create a byte slice of the desired length
	randomBytes := make([]byte, 64)

	_, err := cr.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Convert to a base64 URL-safe string
	return UID(hex.EncodeToString(randomBytes[:])), nil
}
func NewUIDSrc(n int, set string) UID {
	return newBytes(n, set)
}

func (u UID) Bytes() []byte {
	return []byte(u)
}

func (u UID) String() string {
	return string(u)
}

// NewBytes takes letterBytes from parameters and returns random string of length n.
func newBytes(n int, set string) UID {
	bytes := make([]byte, n)
	// A models.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(set) {
			bytes[i] = set[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return UID(bytes)
}
