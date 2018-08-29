package controllers

import (
	"core/context"
	"core/models"
	"core/views"
	"net/http"
	"time"

	"core/rand"
	"fmt"

	"github.com/Sirupsen/logrus"
)

type Users struct {
	LoginView *views.View
	NewView   *views.View
	us        models.UserService
	logger    *logrus.Entry
}

func NewUsers(us models.UserService, logger *logrus.Entry) *Users {
	return &Users{
		LoginView: views.NewView("bootstrap", "users/login"),
		NewView:   views.NewView("bootstrap", "users/new"),
		us:        us,
		logger:    logger,
	}
}

type NewUserForm struct {
	Name             string            `schema:"name"`
	Username         string            `schema:"username"`
	Password         string            `schema:"password"`
	UserRoles        []models.UserRole `schema:"user_roles"`
	UserRolesOptions []models.UserRole `scheme:"user_type_options"`
}

// New to render the form to create new user
// GET /newuser
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	var form NewUserForm
	parseURLParams(r, &form)
	form.UserRolesOptions = models.UserRolesList()
	u.NewView.Render(w, r, form)
}

// Create to process the new user form for creating new user
// POST /newuser
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form NewUserForm
	vd.Yield = &form
	form.UserRolesOptions = models.UserRolesList()
	if err := parseForm(r, &form); err != nil {
		u.logger.Errorln(err)
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}

	user := models.User{
		Name:      form.Name,
		Username:  form.Username,
		Password:  form.Password,
		UserRoles: form.UserRoles,
	}
	if err := u.us.Create(&user); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}
	u.logger.Infoln("User created successfully, redirecting...")
	alert := views.Alert{
		Level:   views.AlertLevelSuccess,
		Message: fmt.Sprintf("User for %s created successfully.", user.Name),
	}

	views.RedirectAlert(w, r, "/admin/dashboard", http.StatusFound, alert)
}

type LoginForm struct {
	Username string `schema:"username"`
	Password string `schema:"password"`
}

// POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	vd := views.Data{}
	form := LoginForm{}
	if err := parseForm(r, &form); err != nil {
		u.logger.Errorf("Error while parsing login form: %v\n", err)
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}
	user, err := u.us.Authenticate(form.Username, form.Password)
	if err != nil {
		switch err.Error() {
		case models.MongoErrNotFound.Error():
			vd.AlertError("Invalid username")
		default:
			vd.SetAlert(err)
		}
		u.LoginView.Render(w, r, vd)
		return
	}

	err = u.signIn(w, user)
	if err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}
	// TODO: redirect to admin overview or by type
	if userRoleExists(models.UserRoleAdmin, user.UserRoles) {
		http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
	} else {
		fmt.Fprintf(w, "Login sucessfull..!! with user: %+v", user)
	}
}

func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {
	if user.Remember == "" {
		token, err := rand.RemeberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		user.LastLogin = time.Now()
		err = u.us.Update(user)
	}
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	return nil
}

// POST /logout
func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	user := context.User(r.Context())
	token, _ := rand.RemeberToken()
	user.Remember = token
	u.us.Update(user)
	http.Redirect(w, r, "/", http.StatusFound)
}

func userRoleExists(role models.UserRole, roles []models.UserRole) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
