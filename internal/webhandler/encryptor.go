package webhandler

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"log"
)

type Encryptor struct {
	aesblock cipher.Block
	key      []byte
}

func newEncryptor(key []byte) *Encryptor {
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal("cannot initialize Encryptor:", err.Error())
	}

	return &Encryptor{key: key, aesblock: aesblock}
}

func (e *Encryptor) Encode(str string) string {
	decoded := []byte(str)
	encoded := make([]byte, e.BlockSize())

	for len(encoded) > len(decoded) {
		decoded = append(decoded, ' ')
	}

	e.aesblock.Encrypt(encoded, decoded)

	return base64.StdEncoding.EncodeToString(encoded)
}

func (e *Encryptor) Decode(str string) (string, error) {
	encoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}

	decoded := make([]byte, e.BlockSize())

	for len(decoded) > len(encoded) {
		encoded = append(encoded, ' ')
	}

	e.aesblock.Decrypt(decoded, []byte(encoded))

	return string(decoded), nil
}

func (e *Encryptor) BlockSize() int {
	return e.aesblock.BlockSize()
}
