package smt

import "go.mongodb.org/mongo-driver/bson/primitive"

// Property struct contains information about the properties to be
// verified
type Property struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	Text        string             `json:"text" bson:"text"`
}
