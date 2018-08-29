package controllers

import (
	"core/views"

	"net/http"

	"core/models"

	"github.com/Sirupsen/logrus"
)

type Admin struct {
	AdminDashboardView *views.View
	logger             *logrus.Entry
	us                 models.UserService
}

func NewAdmin(us models.UserService, logger *logrus.Entry) *Admin {
	return &Admin{
		AdminDashboardView: views.NewView("bootstrap", "admin/dashboard"),
		logger:             logger,
		us:                 us,
	}
}

type AdminDashboardData struct {
	Physicians []models.User
}

// GET /admin/dashboard
func (a *Admin) Dashboard(w http.ResponseWriter, r *http.Request) {
	a.logger.Infoln("Rendering admin dashboard")
	physicians, err := a.us.ByUserRole(models.UserRolePhysician)
	if err != nil {
		a.logger.Errorf("Error while fetching physicians: %+v", err)
		http.Error(w, "Something went wrong while fetching Physicians.", http.StatusInternalServerError)
	}
	a.logger.Debugf("Fetched %d physicians.", len(physicians))
	dashData := AdminDashboardData{}
	dashData.Physicians = physicians
	var vd views.Data
	vd.Yield = dashData
	a.AdminDashboardView.Render(w, r, vd)
}
