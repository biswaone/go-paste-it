package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type store interface {
	PutSnippet(ctx context.Context, id string, snippet *Snippet) error
	GetSnippet(ctx context.Context, id string) (*Snippet, error)
	DeleteSnippet(ctx context.Context, id string) error
	UpdateSnippet(ctx context.Context, id string, snippet *Snippet) error
}

type mongoStore struct {
	client             *mongo.Client
	snippetsCollection *mongo.Collection
}

func newMongoStore(client *mongo.Client) *mongoStore {
	return &mongoStore{
		client:             client,
		snippetsCollection: client.Database("GoPasteIt").Collection("snippets"),
	}
}

func (m *mongoStore) PutSnippet(ctx context.Context, id string, snippet *Snippet) error {
	_, err := m.snippetsCollection.InsertOne(ctx, snippet)
	if err != nil {
		return fmt.Errorf("failed to Insert Snippet: %w", err)
	}
	return nil
}

func (m *mongoStore) GetSnippet(ctx context.Context, id string) (*Snippet, error) {
	filter := bson.M{"id": id}
	var snippet Snippet
	err := m.snippetsCollection.FindOne(ctx, filter).Decode(&snippet)
	if err != nil {
		return nil, err
	}
	return &snippet, nil
}

func (m *mongoStore) DeleteSnippet(ctx context.Context, id string) error {
	_, err := m.snippetsCollection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return fmt.Errorf("failed to Delete snippet: %w", err)
	}
	return nil
}

func (m *mongoStore) UpdateSnippet(ctx context.Context, id string, snippet *Snippet) error {
	filter := bson.M{"id": id}

	update := bson.M{
		"$set": bson.M{
			"title":           snippet.Title,
			"expiration":      snippet.Expiration,
			"burn_after_read": snippet.BurnAfterRead,
			"enable_password": snippet.EnablePassword,
			"password":        snippet.Password,
			"content":         snippet.Content,
			"view_count":      snippet.ViewCount,
		},
	}

	_, err := m.snippetsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to Update snippet: %w", err)
	}
	return nil

}

func (m *mongoStore) createIndexes(ctx context.Context) error {

	// create TTL index on snippet for deletion
	ttlIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "expiration", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}
	_, err := m.snippetsCollection.Indexes().CreateOne(ctx, ttlIndex)
	if err != nil {
		return fmt.Errorf("failed to Create TTL Index: %w", err)
	}

	// create index on id
	idIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = m.snippetsCollection.Indexes().CreateOne(ctx, idIndex)
	if err != nil {
		return fmt.Errorf("failed to Index on Document Id: %w", err)
	}

	return nil
}

func initMongoDB(ctx context.Context, uri string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to verify connection to mongo db: %w", err)
	}

	return client, nil
}
