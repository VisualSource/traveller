package model

import "gopkg.in/mgo.v2/bson"

type Session struct {
	ID    bson.ObjectId
	Name  string
	Admin string
	// TODO other session config stuff here
}
