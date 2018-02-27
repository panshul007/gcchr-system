package models

type Contact struct {
	Email       string `json:"email,omitempty" bson:"email,omitempty"`
	HomePhone   string `json:"home_phone,omitempty" bson:"home_phone,omitempty"`
	OfficePhone string `json:"office_phone,omitempty" bson:"office_phone,omitempty"`
	MobilePhone string `json:"mobile_phone,omitempty" bson:"mobile_phone,omitempty"`
}
