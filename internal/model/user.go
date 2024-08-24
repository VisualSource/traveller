package model

import (
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Username string
	Password string
	//Sessions   []string
	//Characters []string
}
