package utils

import "golang.org/x/crypto/bcrypt"

func CheckPassword(password string, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	return err == nil
}
