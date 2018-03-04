package controllers

import (
	"gcchr-system/core/models"
	"gcchr-system/core/views"
	"net/http"

	"github.com/Sirupsen/logrus"
)

type PatientManagement struct {
	PatientManagementView *views.View
	logger                *logrus.Entry
	ps                    models.PatientService
}

func NewPatientManagement(ps models.PatientService, logger *logrus.Entry) *PatientManagement {
	return &PatientManagement{
		PatientManagementView: views.NewView("bootstrap", "patient/management"),
		logger:                logger,
		ps:                    ps,
	}
}

type PatientManagementData struct {
	Patients      []models.Patient
	TotalPatients int
	Page          int
	PageSize      int
}

// GET /patient/manage
func (pm *PatientManagement) Index(w http.ResponseWriter, r *http.Request) {
	pm.logger.Debugln("Rendering patient management portal")
	patientsPaged, err := pm.ps.AllActive(models.PageInfo{Page: 1, PageSize: 10})
	if err != nil {
		pm.logger.Errorf("Error while fetching patients: %+v", err)
		http.Error(w, "Something went wrong while fetching patients.", http.StatusInternalServerError)
	}
	pm.logger.Debugf("Fetched %d patients\n", len(patientsPaged.Patients))
	portalData := PatientManagementData{
		Patients:      patientsPaged.Patients,
		TotalPatients: patientsPaged.Total,
		Page:          patientsPaged.PageInfo.Page,
		PageSize:      patientsPaged.PageInfo.PageSize,
	}
	var vd views.Data
	vd.Yield = portalData
	pm.PatientManagementView.Render(w, r, vd)
}

// GET /patient/list/:page
func (pm *PatientManagement) PatientListPage(w http.ResponseWriter, r *http.Request) {
	page := 2
	pageSize := 10
	pm.logger.Debugf("Fetching patient list page: %d pageSize: %d\n", page, pageSize)

}
