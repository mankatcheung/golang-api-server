// Package password provides secure password hashing and comparison using bcrypt.
package password

import "golang.org/x/crypto/bcrypt"

const bcryptCost = 12

func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

func Compare(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
