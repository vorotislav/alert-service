package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
)

func GetHash(src, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(src)
	if err != nil {
		return nil, err
	}

	dst := h.Sum(nil)

	return dst, nil
}

func CheckHash(body, hash, key []byte) (bool, error) {
	nh, err := GetHash(body, key)
	if err != nil {
		return false, err
	}

	return bytes.Equal(hash, nh), nil
}
