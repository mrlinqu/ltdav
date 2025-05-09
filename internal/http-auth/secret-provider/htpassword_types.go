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

//--------------

// func randomBytes(n int) ([]byte, error) {
// 	b := make([]byte, n)

// 	r, err := rand.Read(b)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return b[:r], nil
// }

// func randomString(n int) (string, error) {
// 	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

// 	b, err := randomBytes(n)
// 	if err != nil {
// 		return "", err
// 	}

// 	for i, c := range b {
// 		b[i] = letters[c%byte(len(letters))]
// 	}

// 	return string(b), nil
// }

// //==================

// type algBcrypt struct{}

// func (a algBcrypt) IsValid(hash []byte, passwordString []byte) bool {
// 	return bcrypt.CompareHashAndPassword(hash, passwordString) == nil
// }

// func (a algBcrypt) Hash(passwordString []byte) []byte {}

// //==================

// type algSha1 struct{}

// func (a algSha1) IsValid(hash []byte, passwordString []byte) bool {
// 	hashString := a.Hash(passwordString)

// 	return subtle.ConstantTimeCompare(hash, hashString) == 1
// }

// func (a algSha1) Hash(passwordString []byte) []byte {
// 	hh := sha1.New()
// 	hh.Write(passwordString)
// 	return []byte(base64.StdEncoding.EncodeToString(hh.Sum(nil)))
// }

// //==================

// type algMd5 struct{}

// func (a algMd5) IsValid(hash []byte, passwordString []byte) bool {
// 	parts := bytes.SplitN(hash, []byte("$"), 4)
// 	if len(parts) != 4 {
// 		return false
// 	}

// 	magic := []byte("$" + string(parts[1]) + "$")
// 	salt := parts[2]

// 	hashString := md5Crypt(passwordString, salt, magic)

// 	return subtle.ConstantTimeCompare(hash, hashString) == 1
// }

// func (a algMd5) Hash(passwordString []byte) ([]byte, error) {
// 	s, err := randomString(8)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "htpasswd: %v", err)
// 		os.Exit(1)
// 	}

// 	result = password.APR1.Crypt([]byte(passwordString), []byte(s), nil)
// }
