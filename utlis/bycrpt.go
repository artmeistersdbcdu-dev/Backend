package utlis

import "golang.org/x/crypto/bcrypt"

const rounds = 12

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), rounds)
	return string(bytes), err
}

func CheckPassword(password string, realPass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(realPass))
	return err==nil
}
