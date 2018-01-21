package models

import (
	"time"

	"gcchr-system/core/hash"
	"regexp"

	"strings"

	"gcchr-system/core/rand"

	"github.com/Sirupsen/logrus"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"golang.org/x/crypto/bcrypt"
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
	Password     string        `json:"password" bson:"-"`
	PasswordHash string        `json:"password_hash" bson:"password_hash"`
	Remember     string        `json:"remember" bson:"-"`
	RememberHash string        `json:"remember_hash" bson:"remember_hash"`
	Created      time.Time     `json:"created" bson:"created"`
	Updated      time.Time     `json:"updated,omitempty" bson:"updated,omitempty"`
	LastLogin    time.Time     `json:"lastLogin,omitempty" bson:"lastLogin,omitempty"`
	ProfileId    string        `json:"profileId,omitempty" bson:"profileId,omitempty"`
}

type UserDB interface {
	// Single user fetch methods
	ByEmail(email string) (*User, error)
	ById(id string) (*User, error)
	ByRemember(token string) (*User, error)

	// Data modifying methods
	Create(user *User) error
	Update(user *User) error
	Delete(id string) error
}

type userValidator struct {
	UserDB
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
	pepper     string
	logger     *logrus.Entry
}

var _ UserDB = &userValidator{}

func newUserValidator(udb UserDB, logger *logrus.Entry, hmac hash.HMAC, pepper string) *userValidator {
	return &userValidator{
		UserDB:     udb,
		hmac:       hmac,
		emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
		pepper:     pepper,
		logger:     logger,
	}
}

func (uv *userValidator) Create(user *User) error {
	if err := runUserValFuncs(user, uv.passwordRequired, uv.passwordMinLength, uv.bcryptPassword,
		uv.passwordHashRequired, uv.setRememberIfUnset, uv.rememberMinBytes, uv.hmacRemember, uv.rememberHashRequired,
		uv.requireEmail, uv.normalizeEmail, uv.emailFormat, uv.emailIsAvailable); err != nil {
		return err
	}
	return uv.UserDB.Create(user)
}

// Update will update the provided the user with all of the data
// provided in the user object.
func (uv *userValidator) Update(user *User) error {
	if err := runUserValFuncs(user, uv.passwordMinLength, uv.bcryptPassword, uv.passwordHashRequired, uv.rememberMinBytes,
		uv.hmacRemember, uv.rememberHashRequired, uv.normalizeEmail, uv.emailFormat, uv.emailIsAvailable); err != nil {
		return err
	}
	user.Updated = time.Now()
	return uv.UserDB.Update(user)
}

// Delete will delete the user with provided user Id.
func (uv *userValidator) Delete(id string) error {
	if err := uv.isValidId(id); err != nil {
		return err
	}
	return uv.UserDB.Delete(id)
}

func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}
	if err := runUserValFuncs(&user, uv.requireEmail, uv.normalizeEmail); err != nil {
		return nil, err
	}
	return uv.UserDB.ByEmail(user.Email)
}

func (uv *userValidator) ById(id string) (*User, error) {
	if err := uv.isValidId(id); err != nil {
		return nil, err
	}
	return uv.UserDB.ById(id)
}

func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User{
		Remember: token,
	}
	if err := runUserValFuncs(&user, uv.hmacRemember); err != nil {
		return nil, err
	}
	return uv.UserDB.ByRemember(user.RememberHash)
}

func (uv *userValidator) isValidId(id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	}
	return ErrIDInvalid
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func (uv *userValidator) emailFormat(user *User) error {
	if !uv.emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
	}
	return nil
}

func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

func (uv *userValidator) emailIsAvailable(user *User) error {
	existing, err := uv.ByEmail(user.Email)
	if err.Error() == MongoErrNotFound.Error() {
		// Email address is not taken
		return nil
	}
	if err != nil {
		return err
	}

	// we found a user with this email address
	// if the found user has the same ID as this use, it is an update and this is the same user
	if user.Id != existing.Id {
		return ErrEmailTaken
	}
	return nil
}

func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}
	pwByte := []byte(user.Password + uv.pepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwByte, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	return nil
}

func (uv *userValidator) passwordRequired(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}

func (uv *userValidator) passwordHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}
	return nil
}

func (uv *userValidator) passwordMinLength(user *User) error {
	if user.Password == "" {
		return nil
	}
	if len(user.Password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

func (uv *userValidator) setRememberIfUnset(user *User) error {
	if user.Remember != "" {
		return nil
	}

	token, err := rand.RemeberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}
	n, err := rand.NBytes(user.Remember)
	if err != nil {
		return err
	}
	if n < 32 {
		return ErrRememberTokenTooShort
	}
	return nil
}

func (uv *userValidator) rememberHashRequired(user *User) error {
	if user.RememberHash == "" {
		return ErrRememberTokenRequired
	}
	return nil
}

type UserService interface {
	Authenticate(email, password string) (*User, error)
	EnsureAdmin() error
	UserDB
}

type userService struct {
	UserDB
	pepper string
	logger *logrus.Entry
}

func NewUserService(mgo *mgo.Session, logger *logrus.Entry, dbname, pepper, hmacKey string) UserService {
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

func (us *userService) EnsureAdmin() error {
	us.logger.Debugln("Ensuring Admin: admin@gcchr.com")
	email := "admin@gcchr.com"
	u := &User{
		UserType: UserTypeAdmin,
		Email:    email,
		Name:     "GCCHR Admin",
		Password: "adminPass",
		Created:  time.Now(),
	}
	_, err := us.UserDB.ByEmail(email)
	if err != nil {
		us.logger.Debugln("Creating default admin user with email: ", email)
		return us.UserDB.Create(u)
	} else {
		us.logger.Debugln("Admin exists with email: ", email)
	}
	return nil
}

// Authenticate user with provided email and password.
func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password+us.pepper))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrPasswordIncorrect
		default:
			return nil, err
		}
	}
	return foundUser, nil
}

type userMongo struct {
	mgo    *mgo.Session
	dbname string
	logger *logrus.Entry
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

func (um *userMongo) Delete(id string) error {
	ses := um.mgo.Copy()
	defer ses.Close()
	return ses.DB(um.dbname).C(UserCollection).RemoveId(bson.ObjectIdHex(id))
}

func (um *userMongo) Update(user *User) error {
	ses := um.mgo.Copy()
	defer ses.Close()
	return ses.DB(um.dbname).C(UserCollection).UpdateId(user.Id, user)
}

func (um *userMongo) ById(id string) (*User, error) {
	ses := um.mgo.Copy()
	defer ses.Close()
	u := User{}
	err := ses.DB(um.dbname).C(UserCollection).FindId(bson.ObjectIdHex(id)).One(&u)
	return &u, err
}

func (um *userMongo) ByEmail(email string) (*User, error) {
	um.logger.Debugln("Fetching user by email: ", email)
	ses := um.mgo.Copy()
	defer ses.Close()
	u := User{}
	err := ses.DB(um.dbname).C(UserCollection).Find(bson.M{"email": email}).One(&u)
	return &u, err
}

func (um *userMongo) ByRemember(token string) (*User, error) {
	ses := um.mgo.Copy()
	defer ses.Close()
	u := User{}
	err := ses.DB(um.dbname).C(UserCollection).Find(bson.M{"remember_hash": token}).One(&u)
	return &u, err
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
