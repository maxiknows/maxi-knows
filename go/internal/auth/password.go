package auth

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// CheckPassword verifies a password against the stored hash.
// needsUpgrade is true when the hash is a legacy MD5 hash that matched,
// meaning the caller should re-hash with HashPassword and save it.
func CheckPassword(stored, password string) (ok, needsUpgrade bool) {
	// If bcrypt can read a cost from it, it's a bcrypt hash.
	if _, err := bcrypt.Cost([]byte(stored)); err == nil {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil, false
	}

	// Legacy user: old MD5 hash.
	sum := md5.Sum([]byte(password))
	legacy := hex.EncodeToString(sum[:])
	ok = subtle.ConstantTimeCompare([]byte(stored), []byte(legacy)) == 1
	return ok, ok
}
