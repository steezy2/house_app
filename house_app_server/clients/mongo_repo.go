package clients

import (
	"context"
	"log/slog"
	"os"
	"time"

	"house-app/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoRepo holds the database collection
type MongoRepo struct {
	collection *mongo.Collection
}

// NewMongoRepo creates a new repository and connects to the database
func NewMongoRepo() (*MongoRepo, error) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		slog.Error("MONGO_URI environment variable not set")
		os.Exit(1)
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		slog.Error("Failed to connect to MongoDB", "error", err, "uri", mongoURI)
		return nil, err
	}

	dbName := "house_app"
	collectionName := "images"
	db := client.Database(dbName)
	slog.Info("Connected to MongoDB", "database", dbName)

	// Check if the collection exists
	collections, err := db.ListCollectionNames(context.Background(), bson.M{"name": collectionName})
	if err != nil {
		slog.Error("Failed to list collections", "error", err, "database", dbName)
		return nil, err
	}

	collectionExists := false
	for _, name := range collections {
		if name == collectionName {
			collectionExists = true
			break
		}
	}

	if !collectionExists {
		slog.Info("Collection does not exist, creating it", "collection", collectionName)
		err := db.CreateCollection(context.Background(), collectionName)
		if err != nil {
			slog.Error("Failed to create collection", "error", err, "collection", collectionName)
			return nil, err
		}
		slog.Info("Collection created successfully", "collection", collectionName)
	} else {
		slog.Info("Collection already exists", "collection", collectionName)
	}

	collection := db.Collection(collectionName)

	return &MongoRepo{collection: collection}, nil
}

// newImageRecord assigns a fresh ID and creation timestamp, mirroring what
// MongoDB would otherwise leave to CreateImage/CreateImages callers to set
// inconsistently.
func newImageRecord(image models.Image) models.Image {
	image.ID = primitive.NewObjectID()
	image.CreatedAt = time.Now()
	return image
}

// CreateImage creates a new image record in the database
func (r *MongoRepo) CreateImage(image models.Image) (*mongo.InsertOneResult, error) {
	return r.collection.InsertOne(context.Background(), newImageRecord(image))
}

// GetImages retrieves all images from the database
func (r *MongoRepo) GetImages() ([]models.Image, error) {
	var images []models.Image
	cursor, err := r.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &images); err != nil {
		return nil, err
	}
	return images, nil
}

// SearchImages searches for images in the database
func (r *MongoRepo) SearchImages(query string) ([]models.Image, error) {
	filter := bson.M{
		"$text": bson.M{
			"$search": query,
		},
	}

	var images []models.Image
	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &images); err != nil {
		return nil, err
	}
	return images, nil
}

// UpdateImagePath updates the storagePath, category, and processedAt of the
// image record whose storagePath currently equals oldStoragePath. It is a
// no-op (no error) if no record matches, since the processor may retry or
// the record may already have been updated.
func (r *MongoRepo) UpdateImagePath(oldStoragePath, newStoragePath, category string) error {
	filter := bson.M{"storagePath": oldStoragePath}
	update := bson.M{
		"$set": bson.M{
			"storagePath": newStoragePath,
			"category":    category,
			"processedAt": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	return err
}

// CreateImages creates multiple new image records in the database.
func (r *MongoRepo) CreateImages(images []models.Image) (*mongo.InsertManyResult, error) {
	var docs []interface{}
	for _, image := range images {
		docs = append(docs, newImageRecord(image))
	}
	return r.collection.InsertMany(context.Background(), docs)
}

var _ models.ImageRepository = (*MongoRepo)(nil)
