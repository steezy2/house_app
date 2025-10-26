package clients

import (
	"testing"
	"time"

	"house-app/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestMongoRepo_CreateImage(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		repo := &MongoRepo{collection: mt.Coll}
		image := models.Image{
			Filename:    "test.jpg",
			StoragePath: "./uploads/test.jpg",
			Size:        1024,
			ContentType: "image/jpeg",
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		result, err := repo.CreateImage(image)
		if err != nil {
			t.Errorf("CreateImage failed: %v", err)
		}
		if result.InsertedID == nil {
			t.Errorf("Expected InsertedID, got nil")
		}
	})
}

func TestMongoRepo_GetImages(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		repo := &MongoRepo{collection: mt.Coll}
		expectedImages := []models.Image{
			{ID: primitive.NewObjectID(), Filename: "image1.jpg", StoragePath: "./uploads/image1.jpg", CreatedAt: time.Now()},
			{ID: primitive.NewObjectID(), Filename: "image2.png", StoragePath: "./uploads/image2.png", CreatedAt: time.Now()},
		}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: expectedImages[0].ID},
			{Key: "filename", Value: expectedImages[0].Filename},
			{Key: "storagePath", Value: expectedImages[0].StoragePath},
			{Key: "createdAt", Value: expectedImages[0].CreatedAt},
		})
		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, bson.D{
			{Key: "_id", Value: expectedImages[1].ID},
			{Key: "filename", Value: expectedImages[1].Filename},
			{Key: "storagePath", Value: expectedImages[1].StoragePath},
			{Key: "createdAt", Value: expectedImages[1].CreatedAt},
		})
		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, second, killCursors)

		images, err := repo.GetImages()
		if err != nil {
			t.Errorf("GetImages failed: %v", err)
		}
		if len(images) != 2 {
			t.Errorf("Expected 2 images, got %d", len(images))
		}
	})
}
