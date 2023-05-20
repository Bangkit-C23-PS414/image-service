package repository

import (
	"context"
	"image-service/core/domain"
	"io"
	"log"
	"mime/multipart"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/api/option"
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

	firestoreOpt := option.WithCredentialsFile("firestore-sa-key.json")
	firestoreClient, err := firestore.NewClient(ctx, projectId, firestoreOpt)
	if err != nil {
		log.Fatalf("[NewImageRepository] fail to initialize firestore client with error %v \n", err)
		return nil, err
	}

	gcsOpt := option.WithCredentialsFile("gcs-sa-key.json")
	gcsClient, err := storage.NewClient(ctx, gcsOpt)
	if err != nil {
		log.Fatalf("[NewImageRepository] fail to initialize cloud storage client with error %v \n", err)
		return nil, err
	}

	return &ImageRepository{
		firestoreClient: *firestoreClient,
		gcsClient:       *gcsClient,
	}, nil
}

func (i *ImageRepository) UploadImage(username string, file *multipart.File) error {
	ctx := context.Background()
	filename := uuid.New()
	bktName := os.Getenv("CAPSTONE_IMAGE_BUCKET")
	w := i.gcsClient.Bucket(bktName).Object("images/" + filename.String()).NewWriter(ctx)
	_, err := io.Copy(w, *file)
	if err != nil {
		log.Printf("[ImageRepository.UploadImage] error writing to gcs bucket with error %v \n", err)
		return err
	}
	if err = w.Close(); err != nil {
		log.Printf("[ImageRepository.UploadImage] error closing file with error %v \n", err)
		return err
	}

	_, err = i.firestoreClient.Collection("users").Doc(username).Collection("image-detections").Doc(filename.String()).Set(ctx, domain.Image{
		Label:         "",
		InferenceTime: 0,
		UploadedAt:    time.Now(),
	})

	if err != nil {
		log.Printf("[ImageRepository.UploadImage] error write to firestore with error %v \n", err)
		return err
	}
	return nil
}
