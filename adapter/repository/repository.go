package repository

import (
	"context"
	"io"
	"log"
	"mime/multipart"
	"os"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

type ImageRepository struct {
	firestoreClient firestore.Client
	gcsClient       storage.Client
}

func NewImageRepository(ctx context.Context) (*ImageRepository, error) {
	gcsSAFile, err := os.Open("gcs-sa-key.json")
	if err != nil {
		log.Fatalf("[NewImageRepository] unable to open %v with error %v \n", gcsSAFile.Name(), err)
	}
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", gcsSAFile.Name())
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

	// TODO: this code still return code = PermissionDenied desc = Missing or insufficient permissions, fix this later
	_, err = i.firestoreClient.Collection("images").Doc(filename.String()).Create(ctx, map[string]interface{}{
		"Filename": filename.String(),
		"Username": username,
	})
	if err != nil {
		log.Printf("[ImageRepository.UploadImage] error write to firestore with error %v \n", err)
		return err
	}
	return nil
}
