package util

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// MultiCoder can encode/decode depending on which EncFunc or DecFunc
type MultiCoder[T any] struct {
	mu     sync.Mutex
	buff   *bytes.Buffer
	crypto *Crypto
}

func NewMultiCoder[T any]() (*MultiCoder[T], error) {
	block, err := aes.NewCipher(AESKey.Bytes())
	if err != nil {
		return nil, err
	}

	crypto, err := NewCrypto(block)
	if err != nil {
		return nil, err
	}

	return &MultiCoder[T]{
		buff:   bytes.NewBuffer([]byte{}),
		crypto: crypto,
	}, nil
}

// EncFunc is what would be used as argument type for an encoding handler
type EncFunc func(io.Writer, any) error

// DecFunc is what would be used as argument type for a decoding handler
type DecFunc func(io.Reader, any) error

// Encode encodes obj of type T using the EncFunc handler.
func (m *MultiCoder[T]) Encode(obj T, encFunc EncFunc) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.buff.Reset()
	if err := encFunc(m.buff, obj); err != nil {
		return nil, err
	}

	return m.buff.Bytes(), nil
}

// Decode decodes byte array to obj of T using the DecFunc handler.
func (m *MultiCoder[T]) Decode(data []byte, decFunc DecFunc) (T, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.buff.Reset()
	m.buff.Write(data)

	var obj T
	if err := decFunc(m.buff, &obj); err != nil {
		return obj, err
	}

	return obj, nil
}

func (m *MultiCoder[T]) EncodeB64(src T, encFunc EncFunc) ([]byte, error) {
	data, err := m.Encode(src, encFunc)
	if err != nil {
		return nil, err
	}

	m.buff.Reset()
	data, err = m.compress(data)
	if err != nil {
		return nil, err
	}

	//EncodeB64 to B64 UTF8
	dst := make([]byte, base64.URLEncoding.EncodedLen(len(data)))
	base64.URLEncoding.Encode(dst, data)

	// Print
	return dst, nil
}
func (m *MultiCoder[T]) DecodeB64(src []byte, decFunc DecFunc) (T, error) {
	dst := make([]byte, base64.URLEncoding.DecodedLen(len(src)))
	_, err := base64.URLEncoding.Decode(dst, src)
	if err != nil {
		var n T
		return n, err
	}

	dst, err = m.decompress(dst)
	if err != nil {
		var n T
		return n, err
	}

	return m.Decode(dst, decFunc)
}

func (m *MultiCoder[T]) EncodeEncrypt(obj T, encFunc EncFunc) ([]byte, error) {
	data, err := m.Encode(obj, encFunc)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	data, err = m.crypto.EncryptCFB(data)
	m.mu.Unlock()

	if err != nil {
		return nil, err
	}

	return m.compress(data)
}
func (m *MultiCoder[T]) DecodeDecrypt(data []byte, decFunc DecFunc) (T, error) {
	dst, err := m.decompress(data)
	if err != nil {
		var n T
		return n, err
	}
	m.mu.Lock()
	dst, err = m.crypto.DecryptCFB(data)
	m.mu.Unlock()
	if err != nil {
		var n T
		return n, err
	}

	return m.Decode(dst, decFunc)
}

func (m *MultiCoder[T]) EncodeSave(filename string, obj T, encFunc EncFunc) error {
	src, err := m.Encode(obj, encFunc)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(filename), os.ModePerm)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, src, os.ModePerm)
}
func (m *MultiCoder[T]) DecodeOpen(filename string, decFunc DecFunc) (T, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		var n T
		return n, err
	}

	return m.Decode(data, decFunc)
}

func (m *MultiCoder[T]) EncodeChain(obj T, encFunc EncFunc, handlers ...func(T, EncFunc) ([]byte, error)) (data []byte, err error) {
	for _, handler := range handlers {
		data, err = handler(obj, encFunc)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}
func (m *MultiCoder[T]) DecodeChain(data []byte, decFunc DecFunc, handlers ...func([]byte, DecFunc) (T, error)) (obj T, err error) {
	for _, handler := range handlers {
		obj, err = handler(data, decFunc)
		if err != nil {
			return obj, err
		}
	}

	return obj, nil
}

// ====================================================================================================

func (m *MultiCoder[T]) compress(data []byte) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.buff.Reset()

	gz := gzip.NewWriter(m.buff)
	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}

	if err = gz.Close(); err != nil {
		return nil, err
	}

	return m.buff.Bytes(), nil
}
func (m *MultiCoder[T]) decompress(data []byte) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.buff.Reset()

	if len(data) == 0 {
		return nil, io.EOF
	}

	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close() // TODO: Make reusable buffer for this

	_, err = io.Copy(m.buff, gz)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}
	return m.buff.Bytes(), nil
}

func EncodeJson(dst io.Writer, src any) error {
	return json.NewEncoder(dst).Encode(src)
}
func DecodeJson(src io.Reader, dst any) error {
	return json.NewDecoder(src).Decode(dst)
}

func EncodeGob(dst io.Writer, src any) error {
	return gob.NewEncoder(dst).Encode(src)
}
func DecodeGob(src io.Reader, dst any) error {
	return gob.NewDecoder(src).Decode(dst)
}
