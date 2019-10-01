package topology

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
	// FindAll returns all the topology objects from storage
	FindAll() ([]Topology, error)
	// Store stores a new topology object
	Store(t Topology) error
	// Update updates a topology
	Update(t Topology) error
	// DeleteAll deletes all the stored objects
	DeleteAll() error
	// Count counts the topology objects stored
	Count() (int64, error)
}

type repo struct {
	client *mongo.Client
}

// NewRepository returns a new mongo Repository
func NewRepository(c *mongo.Client) Repository {

	return &repo{c}
}

func (r *repo) FindAll() ([]Topology, error) {
	var t []Topology

	collection := r.client.Database("analyzer").Collection("topology")

	ctx := context.Background()
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return t, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var c Topology
		cursor.Decode(&c)
		t = append(t, c)
	}
	if err := cursor.Err(); err != nil {
		log.Println("error getting data from cursor")
		return t, err
	}
	return t, nil
}

func (r *repo) Store(t Topology) error {

	collection := r.client.Database("analyzer").Collection("topology")

	t.ID = primitive.NewObjectID()
	_, err := collection.InsertOne(context.Background(), t)

	if err != nil {
		log.Printf("error storing topology, %v", err)
		return err
	}
	return nil
}

func (r *repo) Update(t Topology) error {
	collection := r.client.Database("analyzer").Collection("topology")

	filter := bson.M{"_id": t.ID}
	update := bson.D{{"$set",
		bson.D{
			{"hosts", t.Hosts},
			{"switches", t.Switches},
			{"links", t.Links},
			{"dot", t.DOT},
			{"dot_img", t.DOTImg},
		},
	}}
	_, err := collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		log.Printf("error updating topology, %v", err)
		return err
	}
	return nil
}

func (r *repo) DeleteAll() error {
	collection := r.client.Database("analyzer").Collection("topology")
	return collection.Drop(nil)
}

func (r *repo) Count() (int64, error) {
	collection := r.client.Database("analyzer").Collection("topology")
	count, err := collection.EstimatedDocumentCount(context.Background())

	if err != nil {
		return -1, err
	}
	return count, nil
}
