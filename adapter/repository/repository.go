package repository

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
)

type ImageRepository struct {
	firestoreClient firestore.Client
	gcsClient       storage.Client
}

func NewImageRepository(ctx context.Context) (*ImageRepository, error) {
	projectId := os.Getenv("CAPSTONE_PROJECT_ID")
	if projectId == "" {
		log.Fatalf("[NewImageRepository] empty project ID")
	}

	firestoreClient, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		log.Fatalf("[NewImageRepository] fail to initialize firestore client with error %v \n", err)
		return nil, err
	}

	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("[NewImageRepository] fail to initialize cloud storage client with error %v \n", err)
		return nil, err
	}

	return &ImageRepository{
		firestoreClient: *firestoreClient,
		gcsClient:       *gcsClient,
	}, nil
}
