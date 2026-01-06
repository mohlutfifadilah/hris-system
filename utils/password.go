// utils/password.go
package utils

import (
	"golang.org/x/crypto/bcrypt"
	"math/rand"
    "time"
)

func HashPassword(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// GenerateRandomPassword generates a random 6-character alphabetic password
func GenerateRandomPassword() string {
    const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    
    rand.Seed(time.Now().UnixNano())
    
    password := make([]byte, 6)
    for i := range password {
        password[i] = letters[rand.Intn(len(letters))]
    }
    
    return string(password)
}