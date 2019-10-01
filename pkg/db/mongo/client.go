package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect connects to a MongoDB instance a returns a mongo client
func Connect(ctx context.Context, URI string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(URI)

	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		return nil, err
	}

	return client, nil
}
