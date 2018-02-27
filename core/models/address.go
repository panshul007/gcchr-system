package models

type AddressType string

const (
	AddressTypeBilling  AddressType = "billing_address"
	AddressTypeDelivery AddressType = "delivery_address"
)

type Address struct {
	AddressType AddressType `json:"address_type" bson:"address_type"`
	FullName    string      `json:"full_name,omitempty" bson:"full_name,omitempty"`
	Street      string      `json:"street" bson:"street"`
	City        string      `json:"city" bson:"city"`
	Pincode     int         `json:"pincode" bson:"pincode"`
	State       string      `json:"state" bson:"state"`
	Country     string      `json:"country" bson:"country"`
}
