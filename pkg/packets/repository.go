package packets

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository defines the methods to be implemented by
// the storage layer.
type Repository interface {
	// FindAll returns all the packets from storage
	FindAll() ([]Packet, error)
	// Store stores a new packet
	Store(p Packet) error
}

type repo struct {
	client *mongo.Client
}

// NewRepository returns a new mongo Repository
func NewRepository(c *mongo.Client) Repository {
	return &repo{c}
}

func (r *repo) FindAll() ([]Packet, error) {

	collection := r.client.Database("analyzer").Collection("packets")

	var packets []Packet

	filter := bson.D{}

	sort := bson.D{{"CreatedAt", 1}}

	options := options.Find()
	options.SetSort(sort)

	cursor, err := collection.Find(context.Background(), filter, options)
	if err != nil {
		log.Println("error getting packets", err.Error())
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var p Packet
		cursor.Decode(&p)
		packets = append(packets, p)
	}
	if err := cursor.Err(); err != nil {
		log.Println("error getting data from cursor", err.Error())
	}
	return packets, nil

}

func (r *repo) Store(p Packet) error {

	collection := r.client.Database("analyzer").Collection("packets")

	p.ID = primitive.NewObjectID()
	_, err := collection.InsertOne(context.Background(), p)

	if err != nil {
		log.Printf("error storing packet, %v", err)
		return err
	}
	return nil
}
