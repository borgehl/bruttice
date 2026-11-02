package util

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

func Sha256Base64String(str string) string {
	h := sha256.Sum256([]byte(str))
	return Base64Encode(h[:])
}

func Sha256Base64Bytes(data []byte) string {
	h := sha256.Sum256(data)
	return Base64Encode(h[:])
}

// Base 64 encode
func Base64Encode(data []byte) string {
	ret := base64.StdEncoding.EncodeToString(data)
	ret = strings.ReplaceAll(ret, "/", "-")
	ret = strings.ReplaceAll(ret, "+", "_")
	ret = strings.ReplaceAll(ret, "=", "")
	return ret
}

// base 64 encode for string
func Base64EncodeString(data string) string {
	return Base64Encode([]byte(data))
}

func Base64DecodeString(encoded string) []byte {
	ret, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic(err)
	}
	return ret
}
