package model

import (
	"fmt"

	"gopkg.in/mgo.v2"
)

type Services struct {
	mgoSession *mgo.Session
	mgo        *mgo.Database
}

func (s *Services) Close() {
	s.mgoSession.Close()
}

func NewServices(configs ...ServicesConfig) (*Services, error) {
	var s Services
	for _, config := range configs {
		if err := config(&s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

type ServicesConfig func(*Services) error

func WithMongoDB(dbConfig DatabaseConfig) ServicesConfig {
	return func(s *Services) error {
		serverAddr := fmt.Sprintf("%s:%d", dbConfig.Host, dbConfig.Port)
		dialInfo := mgo.DialInfo{
			Addrs:    []string{serverAddr},
			Username: dbConfig.User,
			Password: dbConfig.Password,
			Database: dbConfig.Name,
		}
		session, err := mgo.DialWithInfo(&dialInfo)
		if err != nil {
			return err
		}
		s.mgoSession = session
		s.mgo = session.DB(dbConfig.Name)
		return nil
	}
}
