package secret_provider

import (
	"bytes"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

type passwordHash interface {
	IsValid(password []byte) bool
}

type bcryptHash []byte

func (h bcryptHash) IsValid(password []byte) bool {
	return bcrypt.CompareHashAndPassword(h, password) == nil
}

type shaHash []byte

func (h shaHash) IsValid(password []byte) bool {
	hh := sha1.New()
	hh.Write(password)
	hashString := []byte(base64.StdEncoding.EncodeToString(hh.Sum(nil)))

	return subtle.ConstantTimeCompare(h, hashString) == 1
}

type md5Hash []byte

func (h md5Hash) IsValid(password []byte) bool {
	parts := bytes.SplitN(h, []byte("$"), 4)
	if len(parts) != 4 {
		return false
	}

	magic := []byte("$" + string(parts[1]) + "$")
	salt := parts[2]

	hashString := md5Crypt(password, salt, magic)

	return subtle.ConstantTimeCompare(h, hashString) == 1
}
