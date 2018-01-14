package model

import (
	"time"

	"gcchr-system/core/hash"
	"regexp"

	"github.com/Sirupsen/logrus"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
)

const (
	UserCollection = "user"
	UserTypeAdmin  = "admin"
)

type User struct {
	Id           bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	UserType     string        `json:"user_type" bson:"user_type"`
	Name         string        `json:"name" bson:"name"`
	Email        string        `json:"email" bson:"email"`
	Password     string        `json:"password"`
	PasswordHash string        `json:"password_hash" bson:"password_hash"`
	Remember     string        `json:"remember"`
	RememberHash string        `json:"remember_hash" bson:"remember_hash"`
	Created      time.Time     `json:"created" bson:"created"`
	Updated      time.Time     `json:"updated,omitempty" bson:"updated,omitempty"`
	LastLogin    time.Time     `json:"lastLogin,omitempty" bson:"lastLogin,omitempty"`
	ProfileId    uuid.UUID     `json:"profileId,omitempty" bson:"profileId,omitempty"`
}

type UserDB interface {
	// Single user fetch methods

	// Data modifying methods
	Create(user *User) error
}

type userValidator struct {
	UserDB
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
	pepper     string
	logger     *logrus.Logger
}

var _ UserDB = &userValidator{}

func newUserValidator(udb UserDB, logger *logrus.Logger, hmac hash.HMAC, pepper string) *userValidator {
	return &userValidator{
		UserDB:     udb,
		hmac:       hmac,
		emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
		pepper:     pepper,
		logger:     logger,
	}
}

func (uv *userValidator) Create(user *User) error {
	uv.logger.Infoln("Validating user with email: ", user.Email)
	return uv.UserDB.Create(user)
}

type UserService interface {
	//Authenticate(email, password string) (*User, error)
	EnsureAdmin()
	UserDB
}

type userService struct {
	UserDB
	pepper string
	logger *logrus.Logger
}

func NewUserService(mgo *mgo.Session, logger *logrus.Logger, dbname, pepper, hmacKey string) UserService {
	um := &userMongo{mgo, dbname, logger}
	hmac := hash.NewHMAC(hmacKey)
	uv := newUserValidator(um, logger, hmac, pepper)

	// Returns an instance of UserService which calls its methods from UserDB which is actually an instance of
	// userValidator, which in turn calls its methods of UserDB which is actually an instance of um.
	return &userService{
		UserDB: uv,
		pepper: pepper,
		logger: logger,
	}
}

func (us *userService) EnsureAdmin() {
	us.logger.Infoln("Ensuring Admin: admin@gcchr.com")
	u := &User{
		UserType: UserTypeAdmin,
		Email:    "admin@gcchr.com",
		Password: "adminPass",
		Created:  time.Now(),
	}
	us.UserDB.Create(u)
}

type userMongo struct {
	mgo    *mgo.Session
	dbname string
	logger *logrus.Logger
}

// To ensure that userMongo is implementing UserDB interface
// if at any point this is not true, we will get a compilation error.
var _ UserDB = &userMongo{}

func (um *userMongo) Create(user *User) error {
	um.logger.Infoln("creating user with email: ", user.Email)
	ses := um.mgo.Copy()
	defer ses.Close()
	return ses.DB(um.dbname).C(UserCollection).Insert(user)
}

type userValFunc func(user *User) error

func runUserValFuncs(user *User, fns ...userValFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}
