package passwordhasher

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher struct{}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}

func (b *PasswordHasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", errors.Wrap(err, "passwordhasher: PasswordHasher.HashPassword bcrypt.GenerateFromPassword error")
	}

	return string(bytes), nil
}

func (b *PasswordHasher) VerifyPassword(password string, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return errors.Wrap(err, "passwordhasher: PasswordHasher.VerifyPassword bcrypt.CompareHashAndPassword error")
	}

	return nil
}
