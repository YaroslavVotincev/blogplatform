package cryptservice

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func CryptValue(value string) (string, error) {
	password, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("fail to crypt value cause %v", err)
	}
	return string(password), nil
}

//func ValueHashMatched(value string, hashedValue string) bool {
//	return bcrypt.CompareHashAndPassword([]byte(hashedValue), []byte(value)) == nil
//}
