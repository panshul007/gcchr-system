package main

import (
	"flag"
	"fmt"
	"gcchr-system/core/model"
)

func main() {

	prodEnv := flag.Bool("prod", false, "Set to true to run the server in production mode. core.config is required if set to true.")

	flag.Parse()
	config := model.LoadConfig(*prodEnv)

	fmt.Println("This is gcchr system core.")
	fmt.Printf("%+v\n", config)
	user := model.User{}
	user.Create()
}
