package smt

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Repository defines the methods to be implemented by
// the storage layer.
type Repository interface {
	// FindAll returns all the packets from storage
	FindAll() ([]Property, error)
	// Store stores a new packet
	Store(p Property) error
	// Update updates a property
	Update(t Property) error
	// Count counts the properties objects stored
	Count() (int64, error)
}

type repo struct {
	client *mongo.Client
}

// NewRepository returns a new mongo Repository
func NewRepository(c *mongo.Client) Repository {
	return &repo{c}
}

func (r *repo) Store(p Property) error {
	collection := r.client.Database("analyzer").Collection("smt")

	p.ID = primitive.NewObjectID()
	_, err := collection.InsertOne(context.Background(), p)

	if err != nil {
		log.Printf("error storing properties, %v", err)
		return err
	}
	return nil
}

func (r *repo) FindAll() ([]Property, error) {
	var t []Property

	collection := r.client.Database("analyzer").Collection("smt")

	ctx := context.Background()
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return t, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var c Property
		cursor.Decode(&c)
		t = append(t, c)
	}
	if err := cursor.Err(); err != nil {
		log.Println("error getting data from cursor")
		return t, err
	}
	return t, nil
}

func (r *repo) Update(p Property) error {
	collection := r.client.Database("analyzer").Collection("smt")

	filter := bson.M{"_id": p.ID}
	update := bson.D{{"$set",
		bson.D{
			{"title", p.Title},
			{"description", p.Description},
			{"text", p.Text},
		},
	}}
	_, err := collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		log.Printf("error updating property, %v", err)
		return err
	}
	log.Printf("updated property %v", update)
	return nil
}

func (r *repo) Count() (int64, error) {
	collection := r.client.Database("analyzer").Collection("smt")
	count, err := collection.EstimatedDocumentCount(context.Background())

	if err != nil {
		return -1, err
	}
	return count, nil
}
