package views

import (
	"log"
	"net/http"
	"time"

	"lenslocked.com/models"
)

const (
	AlertLevelError   = "danger"
	AlertLevelWarning = "warning"
	AlertLevelInfo    = "info"
	AlertLevelSuccess = "success"

	AlertMessageGeneric = "Something went wrong. Please try again, and contact us if the problem persists."
)

// Alert is used to render Bootstrap Alert messages in templates
type Alert struct {
	Level   string
	Message string
}

// Data is the top level structure that views expect data to be passed in.
type Data struct {
	Alert *Alert
	User  *models.User
	Yield interface{}
}

func (d *Data) SetAlert(err error) {
	if pErr, ok := err.(PublicError); ok { // Type casting error into Public error, if succeeds, the ok is set true and pErr has the type casted error.
		d.Alert = &Alert{
			Level:   AlertLevelError,
			Message: pErr.Public(),
		}
	} else {
		log.Println(err)
		d.Alert = &Alert{
			Level:   AlertLevelError,
			Message: AlertMessageGeneric,
		}
	}
}

func (d *Data) AlertError(msg string) {
	d.Alert = &Alert{
		Level:   AlertLevelError,
		Message: msg,
	}
}

type PublicError interface {
	error
	Public() string
}

func persistAlert(w http.ResponseWriter, alert Alert) {
	expiresAt := time.Now().Add(5 * time.Minute)
	lvl := http.Cookie{
		Name:     "alert_level",
		Value:    alert.Level,
		Expires:  expiresAt,
		HttpOnly: true,
	}
	msg := http.Cookie{
		Name:     "alert_message",
		Value:    alert.Message,
		Expires:  expiresAt,
		HttpOnly: true,
	}
	http.SetCookie(w, &lvl)
	http.SetCookie(w, &msg)
}

func clearAlert(w http.ResponseWriter) {
	lvl := http.Cookie{
		Name:     "alert_level",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	msg := http.Cookie{
		Name:     "alert_message",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	http.SetCookie(w, &lvl)
	http.SetCookie(w, &msg)
}

func getAlert(r *http.Request) *Alert {
	lvl, err := r.Cookie("alert_level")
	if err != nil {
		return nil
	}
	msg, err := r.Cookie("alert_message")
	if err != nil {
		return nil
	}
	alert := Alert{
		Level:   lvl.Value,
		Message: msg.Value,
	}
	return &alert
}

func RedirectAlert(w http.ResponseWriter, r *http.Request, urlStr string, code int, alert Alert) {
	persistAlert(w, alert)
	http.Redirect(w, r, urlStr, code)
}
