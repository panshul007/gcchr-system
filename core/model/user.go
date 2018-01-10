package model

import "fmt"

type User struct {
	UserType string
}

func (u *User) Create() error {
	fmt.Println("This is user create method.")
	return nil
}