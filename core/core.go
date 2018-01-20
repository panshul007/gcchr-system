package main

import (
	"flag"
	"fmt"
	"gcchr-system/core/models"
)

func main() {

	prodEnv := flag.Bool("prod", false, "Set to true to run the server in production mode. core.config is required if set to true.")

	flag.Parse()
	config := models.LoadConfig(*prodEnv)

	fmt.Println("This is gcchr system core.")
	fmt.Printf("%+v\n", config)

	services, err := models.NewServices(
		models.WithLogger(config.LogConfig),
		models.WithMongoDB(config.MongoDB),
		models.WithUserService(config.Pepper, config.HMACKey),
	)
	must(err)
	defer services.Close()
	ensureAdmin(services.User)
}

func ensureAdmin(us models.UserService) {
	fmt.Println("Ensuring admin")
	err := us.EnsureAdmin()
	must(err)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
