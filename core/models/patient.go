package models

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	PatientCollection                  = "patient"
	PatientStatusActive  PatientStatus = "active"
	PatientStatusDeleted PatientStatus = "deleted"
)

type PatientStatus string

type Patient struct {
	Id          bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string        `json:"name" bson:"name"`
	DateOfBirth time.Time     `json:"date_of_birth" bson:"date_of_birth"`
	GcchrId     string        `json:"gcchr_id" bson:"gcchr_id"`
	Created     time.Time     `json:"created" bson:"created"`
	Updated     time.Time     `json:"updated,omitempty" bson:"updated,omitempty"`
	Contact     Contact       `json:"contact,omitempty" bson:"contact,omitempty"`
	Addresses   []Address     `json:"addresses,omitempty" bson:"addresses,omitempty"`
	Status      PatientStatus `json:"status" bson:"status"`
}

type PagedResponse struct {
	Patients []Patient
	Total    int
	Page     int
	PageSize int
}

type PatientDB interface {
	// Single patient fetch methods
	ById(id string) (*Patient, error)

	// List of patients fetch methods
	ByName(name string, pageInfo PageInfo) (*PagedResponse, error)
	ByStatus(status PatientStatus, pageInfo PageInfo) (*PagedResponse, error)
	All(pageInfo PageInfo) (*PagedResponse, error)

	// Data modifying methods
	Create(patient *Patient) error
	Update(patient *Patient) error
	Delete(patient *Patient) error
}

type patientValidator struct {
	PatientDB
	logger *logrus.Entry
}

var _ PatientDB = &patientValidator{}

func newPatientValidator(pdb PatientDB, logger *logrus.Entry) *patientValidator {
	return &patientValidator{
		PatientDB: pdb,
		logger:    logger,
	}
}

type PatientService interface {
	EnsurePatient() error
	PatientDB
}

type patientService struct {
	PatientDB
	logger *logrus.Entry
}

func NewPatientService(mgo *mgo.Session, logger *logrus.Entry, dbname string) PatientService {
	pm := &patientMongo{mgo, dbname, logger}
	pv := newPatientValidator(pm, logger)

	return &patientService{
		PatientDB: pv,
		logger:    logger,
	}
}

func (ps *patientService) EnsurePatient() error {
	return nil
}

type patientMongo struct {
	mgo    *mgo.Session
	dbname string
	logger *logrus.Entry
}

var _ PatientDB = &patientMongo{}

func (pm *patientMongo) Create(patient *Patient) error {
	pm.logger.Debugln("creating patient with name: ", patient.Name)
	ses := pm.mgo.Copy()
	defer ses.Close()
	return ses.DB(pm.dbname).C(PatientCollection).Insert(patient)
}

func (pm *patientMongo) Update(patient *Patient) error {
	ses := pm.mgo.Copy()
	defer ses.Close()
	return ses.DB(pm.dbname).C(PatientCollection).UpdateId(patient.Id, patient)
}

func (pm *patientMongo) Delete(id string) error {
	p, err := pm.ById(id)
	if err != nil {
		return err
	}
	p.Status = PatientStatusDeleted
	return pm.Update(p)
}

func (pm *patientMongo) ById(id string) (*Patient, error) {
	ses := pm.mgo.Copy()
	defer ses.Close()
	p := Patient{}
	err := ses.DB(pm.dbname).C(PatientCollection).FindId(bson.ObjectIdHex(id)).One(&p)
	return &p, err
}
