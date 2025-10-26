package clients

import (
	"context"
	"log"
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
		log.Fatal("MONGO_URI environment variable not set")
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	dbName := "house_app"
	collectionName := "images"
	db := client.Database(dbName)
	log.Printf("Connected to MongoDB and using database '%s'", dbName)

	// Check if the collection exists
	collections, err := db.ListCollectionNames(context.Background(), bson.M{"name": collectionName})
	if err != nil {
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
		log.Printf("Collection '%s' does not exist, creating it...", collectionName)
		err := db.CreateCollection(context.Background(), collectionName)
		if err != nil {
			return nil, err
		}
		log.Printf("Collection '%s' created successfully.", collectionName)
	} else {
		log.Printf("Collection '%s' already exists.", collectionName)
	}

	collection := db.Collection(collectionName)

	return &MongoRepo{collection: collection}, nil
}

// CreateImage creates a new image record in the database
func (r *MongoRepo) CreateImage(image models.Image) (*mongo.InsertOneResult, error) {
	image.ID = primitive.NewObjectID()
	image.CreatedAt = time.Now()
	return r.collection.InsertOne(context.Background(), image)
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

// CreateImages creates multiple new image records in the database.
func (r *MongoRepo) CreateImages(images []models.Image) (*mongo.InsertManyResult, error) {
	var docs []interface{}
	for _, image := range images {
		image.ID = primitive.NewObjectID()
		image.CreatedAt = time.Now()
		docs = append(docs, image)
	}
	return r.collection.InsertMany(context.Background(), docs)
}

var _ models.ImageRepository = (*MongoRepo)(nil)
