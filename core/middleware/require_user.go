package middleware

import (
	"gcchr-system/core/context"
	"gcchr-system/core/models"
	"net/http"
	"strings"
)

type User struct {
	models.UserService
}

func (u *User) Apply(next http.Handler) http.HandlerFunc {
	return u.ApplyFunc(next.ServeHTTP)
}

func (u *User) ApplyFunc(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// User lookup not required for static pages and public assets
		path := r.URL.Path
		if strings.HasPrefix(path, "/assets") ||
			strings.HasPrefix(path, "/images") {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("remember_token")
		if err != nil {
			next(w, r)
			return
		}
		user, err := u.UserService.ByRemember(cookie.Value)
		if err != nil {
			next(w, r)
			return
		}
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)
		next(w, r)

	})
}

type RequireUser struct {
	User
}

// Apply assumes that User middleware has already been run.
func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFunc(next.ServeHTTP)
}

// ApplyFunc assumes that User middleware has already been run.
func (mw *RequireUser) ApplyFunc(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	})

}
