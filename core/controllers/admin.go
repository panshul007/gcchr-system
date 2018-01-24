package controllers

import (
	"gcchr-system/core/views"

	"net/http"

	"github.com/Sirupsen/logrus"
)

type Admin struct {
	AdminDashboardView *views.View
	logger             *logrus.Entry
}

func NewAdmin(logger *logrus.Entry) *Admin {
	return &Admin{
		AdminDashboardView: views.NewView("bootstrap", "admin/dashboard"),
		logger:             logger,
	}
}

// GET /admin/dashboard
func (a *Admin) Dashboard(w http.ResponseWriter, r *http.Request) {
	a.logger.Infoln("Rendering admin dashboard")
	var vd views.Data
	a.AdminDashboardView.Render(w, r, vd)
}
