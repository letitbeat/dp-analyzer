package topology

import "go.mongodb.org/mongo-driver/bson/primitive"

// Topology represents a data-plane topology
type Topology struct {
	ID       primitive.ObjectID `json:"id" bson:"_id"`
	Hosts    []string           `json:"hosts" bson:"hosts"`
	Switches []string           `json:"switches" bson:"switches"`
	Links    []string           `json:"links" bson:"links"`
	DOT      string             `json:"dot" bson:"dot"`
	DOTImg   string             `json:"dot_img" bson:"dot_img"`
}
