package tools

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func ComputeSign(bytes []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(bytes)
	dst := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(dst)
}

func AreEquals(sign1 string, sign2 string) bool {
	s1, _ := base64.StdEncoding.DecodeString(sign1)
	s2, _ := base64.StdEncoding.DecodeString(sign2)
	return hmac.Equal(s1, s2)
}
