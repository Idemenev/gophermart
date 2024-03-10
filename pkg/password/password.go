package password

import "golang.org/x/crypto/bcrypt"

const PasswordMaxLength = 72 // ограничение bcrypt.GenerateFromPassword

func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CompareHashAndPassword(passwordHash string, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(plainPassword))
	return err == nil
}
