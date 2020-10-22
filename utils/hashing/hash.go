package hashing

import (
	"crypto/md5"
	"encoding/hex"
)

func CalculateMD5Hash(text string) string {
	hasher := md5.New()
	_, err := hasher.Write([]byte(text))
	if err != nil {
		return ""
	}
	return hex.EncodeToString(hasher.Sum(nil))
}
