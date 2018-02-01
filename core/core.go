package main

import (
	"flag"
	"fmt"
	"gcchr-system/core/controllers"
	"gcchr-system/core/middleware"
	"gcchr-system/core/models"
	"net/http"

	"github.com/gorilla/mux"
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

	r := mux.NewRouter()
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(services.User, services.GetContextLogger("UserController"))
	adminC := controllers.NewAdmin(services.GetContextLogger("AdminController"))

	//b, err := rand.Bytes(32)
	must(err)
	//csrfMw := csrf.Protect(b, csrf.Secure(config.IsProd()))
	userMw := middleware.User{UserService: services.User}
	requireUserMw := middleware.RequireUser{User: userMw}

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/logout", requireUserMw.ApplyFunc(usersC.Logout)).Methods("POST")

	// Admin
	r.HandleFunc("/admin/dashboard", requireUserMw.ApplyFunc(adminC.Dashboard)).Methods("GET")
	r.HandleFunc("/newuser", requireUserMw.ApplyFunc(usersC.New)).Methods("GET")
	r.HandleFunc("/newuser", requireUserMw.ApplyFunc(usersC.Create)).Methods("POST")

	// Assets
	assetHandler := http.FileServer(http.Dir("./core/assets"))
	assetHandler = http.StripPrefix("/assets/", assetHandler)
	r.PathPrefix("/assets/").Handler(assetHandler)

	fmt.Printf("Starting the server at port :%d...\n", config.Port)
	// To apply the user middleware to all requests received.
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), userMw.Apply(r))
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
