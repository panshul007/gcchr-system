package main

import (
	"fmt"
	"gcchr-system/core/model"
)

func main() {
	fmt.Println("This is gcchr system core.")
	user := model.User{}
	user.Create()
}
