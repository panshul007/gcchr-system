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

	services, err := model.NewServices(
		model.WithLogger(config.LogConfig),
		model.WithMongoDB(config.MongoDB),
		model.WithUserService(config.Pepper, config.HMACKey),
	)
	must(err)
	defer services.Close()
	ensureAdmin(services.User)
}

func ensureAdmin(us model.UserService) {
	fmt.Println("Ensuring admin")
	err := us.EnsureAdmin()
	must(err)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
