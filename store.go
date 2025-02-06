package main

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type store interface {
	PutSnippet(ctx context.Context, id string, snippet *Snippet) error
	GetSnippet(ctx context.Context, id string) (*Snippet, error)
	DeleteSnippet(ctx context.Context, id string) error
}

type mongoStore struct {
	client *mongo.Client
}

func (m *mongoStore) PutSnippet(ctx context.Context, id string, snippet *Snippet) error {
	collection := m.client.Database("snippets").Collection("snippets")
	_, err := collection.InsertOne(ctx, snippet)
	return err
}

func (m *mongoStore) GetSnippet(ctx context.Context, id string) (*Snippet, error) {
	collection := m.client.Database("snippets").Collection("snippets")
	filter := bson.M{"id": id}

	var snippet Snippet
	err := collection.FindOne(ctx, filter).Decode(&snippet)
	if err != nil {
		return nil, err
	}
	return &snippet, nil
}

func (m *mongoStore) DeleteSnippet(ctx context.Context, id string) error {
	collection := m.client.Database("snippets").Collection("snippets")
	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func initMongoDB(uri string, ctx context.Context) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}
