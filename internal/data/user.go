package data

import (
	"errors"
	"time"

	"github.com/hayohtee/go-backend-template/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrDuplicateEmail is a custom error that is return when a user
	// try to register with an existing email.
	ErrDuplicateEmail = errors.New("duplicate email")
)

// User is a struct to represent an individual user.
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// password is a struct containing the plaintext and hashed
// versions of the password for a user.
//
// The plaintext is a pointer to a string to help distinguish
// between a plaintext password not being present in the struct
// versus a plaintext password which is empty string.
type password struct {
	plaintext *string
	hash      []byte
}

// Set calculates the bycrypt hash of plaintextPassword and stores
// both the hash and the plaintext versions in the struct.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// Matches checks whether the provided plaintext password matches the
// hashed password stored in the struct, returning true if it matches
// and false otherwise.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePassword(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)
	if user.Password.plaintext != nil {
		ValidatePassword(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing hash for user")
	}
}
