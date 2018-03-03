package models

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
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
	Gender      string        `json:"gender" bson:"gender"`
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
	PageInfo PageInfo
}

type PatientDB interface {
	// Single patient fetch methods
	ById(id string) (*Patient, error)

	// List of patients fetch methods
	ByName(name string, pageInfo PageInfo) (*PagedResponse, error)
	ByStatus(status PatientStatus, pageInfo PageInfo) (*PagedResponse, error)
	All(pageInfo PageInfo) (*PagedResponse, error)
	AllActive(pageInfo PageInfo) (*PagedResponse, error)

	// Data modifying methods
	Create(patient *Patient) error
	Update(patient *Patient) error
	Delete(id string) error
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

func (pv *patientValidator) Create(patient *Patient) error {
	if err := runPatientValFuncs(patient, pv.requireName, pv.ensureGcchrId, pv.ensureCreatedAt, pv.ensureStatus); err != nil {
		return err
	}
	return pv.PatientDB.Create(patient)
}

func (pv *patientValidator) Update(patient *Patient) error {
	if err := runPatientValFuncs(patient, pv.ensureGcchrId, pv.ensureUpdatedAt); err != nil {
		return err
	}
	return pv.PatientDB.Update(patient)
}

func (pv *patientValidator) Delete(id string) error {
	if err := pv.isValidId(id); err != nil {
		return err
	}
	return pv.PatientDB.Delete(id)
}

func (pv *patientValidator) isValidId(id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	}
	return ErrIDInvalid
}

func (pv *patientValidator) requireName(patient *Patient) error {
	if patient.Name == "" {
		return ErrPatientNameRequired
	}
	return nil
}

func (pv *patientValidator) ensureGcchrId(patient *Patient) error {
	if patient.GcchrId == "" {
		patient.GcchrId = uuid.New().String()
	}
	return nil
}

func (pv *patientValidator) ensureCreatedAt(patient *Patient) error {
	if patient.Created.IsZero() {
		patient.Created = time.Now()
	}
	return nil
}

func (pv *patientValidator) ensureUpdatedAt(patient *Patient) error {
	patient.Updated = time.Now()
	return nil
}

func (pv *patientValidator) ensureStatus(patient *Patient) error {
	if patient.Status == "" {
		patient.Status = PatientStatusActive
	}
	return nil
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

// TODO: validate pageInfo
func (pm *patientMongo) All(pageInfo PageInfo) (*PagedResponse, error) {
	query := bson.M{}
	return fetchPagedResultForQuery(pm, query, pageInfo)
}

func (pm *patientMongo) ByName(name string, pageInfo PageInfo) (*PagedResponse, error) {
	pm.logger.Debugln("Fetching patients with name: " + name)
	query := bson.M{"name": name, "status": PatientStatusActive}
	return fetchPagedResultForQuery(pm, query, pageInfo)
}

func (pm *patientMongo) ByStatus(status PatientStatus, pageInfo PageInfo) (*PagedResponse, error) {
	pm.logger.Debugln("Fetching all patients with status: " + status)
	query := bson.M{"status": status}
	return fetchPagedResultForQuery(pm, query, pageInfo)
}

func (pm *patientMongo) AllActive(pageInfo PageInfo) (*PagedResponse, error) {
	pm.logger.Debugln("Fetching all active patients")
	query := bson.M{"status": PatientStatusActive}
	return fetchPagedResultForQuery(pm, query, pageInfo)
}

func fetchPagedResultForQuery(pm *patientMongo, query bson.M, pageInfo PageInfo) (*PagedResponse, error) {
	ses := pm.mgo.Copy()
	defer ses.Close()
	skips := pageInfo.PageSize * (pageInfo.Page - 1)
	var patients []Patient
	err := ses.DB(pm.dbname).C(PatientCollection).Find(query).Skip(skips).Limit(pageInfo.PageSize).All(&patients)
	if err != nil {
		return nil, err
	}
	total, err := ses.DB(pm.dbname).C(PatientCollection).Find(query).Count()
	if err != nil {
		return nil, err
	}
	return &PagedResponse{
		Patients: patients,
		Total:    total,
		PageInfo: pageInfo,
	}, nil
}

type patientValFunc func(patient *Patient) error

func runPatientValFuncs(patient *Patient, fns ...patientValFunc) error {
	for _, fn := range fns {
		if err := fn(patient); err != nil {
			return err
		}
	}
	return nil
}
