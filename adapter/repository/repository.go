package repository

import (
	"context"
	"fmt"
	"image-service/core/domain"
	"io"
	"log"
	"mime/multipart"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
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

	_, err = i.firestoreClient.Collection("images").Doc(filename.String()).Set(ctx, domain.Image{
		Username:   username,
		Filename:   filename.String(),
		UploadedAt: time.Now(),
	})

	if err != nil {
		log.Printf("[ImageRepository.UploadImage] error write to firestore with error %v \n", err)
		return err
	}

	return nil
}

func (i *ImageRepository) GetDetectionResults(username string) ([]domain.Image, error) {
	bktName := os.Getenv("CAPSTONE_IMAGE_BUCKET")
	gcsOpt := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(7 * 24 * time.Hour),
	}

	result := []domain.Image{}
	q := i.firestoreClient.Collection("images").Where("username", "==", username).Documents(context.Background())
	for {
		doc, err := q.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		objectUrl, err := i.gcsClient.Bucket(bktName).SignedURL(fmt.Sprintf("images/%v", doc.Data()["filename"]), gcsOpt)

		if err != nil {
			log.Printf("[ImageRepository.GetDetectionResults] error generate signed URL with error %v \n", err)
			return nil, err
		}

		data := domain.Image{
			Username:      fmt.Sprint(doc.Data()["username"]),
			Filename:      objectUrl,
			Label:         fmt.Sprint(doc.Data()["label"]),
			InferenceTime: doc.Data()["inferenceTime"].(int64),
			UploadedAt:    doc.Data()["uploadedAt"].(time.Time),
			DetectedAt:    doc.Data()["detectedAt"].(time.Time),
		}
		result = append(result, data)
	}
	return result, nil
}

func (i *ImageRepository) UpdateImageResult(payload domain.UpdateImagePayload) error {
	ctx := context.Background()
	_, err := i.firestoreClient.Collection("images").Doc(payload.Filename).Update(ctx, []firestore.Update{
		{
			Path:  "detectedAt",
			Value: time.Now(),
		},
		{
			Path:  "inferenceTime",
			Value: payload.InferenceTime,
		},
		{
			Path:  "label",
			Value: payload.Label,
		},
	})

	if err != nil {
		log.Printf("[ImageRepository.UpdateImageResult] error when update image result with error %v", err)
		return err
	}
	return nil
}
