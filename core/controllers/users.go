package controllers

import (
	"gcchr-system/core/context"
	"gcchr-system/core/models"
	"gcchr-system/core/views"
	"net/http"
	"time"

	"fmt"
	"gcchr-system/core/rand"

	"github.com/Sirupsen/logrus"
)

type Users struct {
	LoginView *views.View
	us        models.UserService
	logger    *logrus.Entry
}

func NewUsers(us models.UserService, logger *logrus.Entry) *Users {
	return &Users{
		LoginView: views.NewView("bootstrap", "users/login"),
		us:        us,
		logger:    logger,
	}
}

type LoginForm struct {
	Email    string `schema:"email"`
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
	user, err := u.us.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.MongoErrNotFound:
			vd.AlertError("Invalid email address")
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
	fmt.Fprintf(w, "Login sucessfull..!! with user: %+v", user)
}

func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {
	if user.Remember == "" {
		token, err := rand.RemeberToken()
		if err != nil {
			return err
		}
		user.Remember = token
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
