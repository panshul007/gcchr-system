package models

import "strings"

const (
	// ErrNotFound Error returned when resource not found.
	ErrNotFound          modelError = "models: resource not found"
	ErrPasswordIncorrect modelError = "models: incorrect password provided"
	ErrEmailRequired     modelError = "models: email address is required"
	ErrEmailInvalid      modelError = "models: email address is not valid"
	ErrEmailTaken        modelError = "models: email address is already taken"
	ErrPasswordTooShort  modelError = "models: password must be at least 8 characters log"
	ErrPasswordRequired  modelError = "models: password is required"
	ErrTitleRequired     modelError = "models: title is required"

	ErrIDInvalid             privateError = "models: ID provided was invalid"
	ErrRememberTokenTooShort privateError = "models: remember token should be at least 32 bytes"
	ErrRememberTokenRequired privateError = "models: remember token is required"
	ErrUserIDRequired        privateError = "models: user ID is required"

	MongoErrNotFound mongoError = "not found"
)

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	split := strings.Split(s, " ")
	split[0] = strings.Title(split[0])
	return strings.Join(split, " ")
}

type privateError string

func (e privateError) Error() string {
	return string(e)
}

type mongoError string

func (e mongoError) Error() string {
	return string(e)
}
