package model

import "gopkg.in/mgo.v2/bson"

type Character struct {
	ID    bson.ObjectId
	Owner string
	// Other character info
}
