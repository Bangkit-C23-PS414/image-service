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

func (i *ImageRepository) UploadImage(email string, file *multipart.File) (*domain.Image, error) {
	ctx := context.Background()
	filename := uuid.New()
	bktName := os.Getenv("CAPSTONE_IMAGE_BUCKET")
	w := i.gcsClient.Bucket(bktName).Object("images/" + filename.String()).NewWriter(ctx)
	_, err := io.Copy(w, *file)
	if err != nil {
		log.Printf("[ImageRepository.UploadImage] error writing to gcs bucket with error %v \n", err)
		return nil, err
	}
	if err = w.Close(); err != nil {
		log.Printf("[ImageRepository.UploadImage] error closing file with error %v \n", err)
		return nil, err
	}

	gcsOpt := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(7 * 24 * time.Hour),
	}
	objectUrl, err := i.gcsClient.Bucket(bktName).SignedURL(fmt.Sprintf("images/%v", filename.String()), gcsOpt)
	if err != nil {
		return nil, err
	}

	data := domain.Image{
		Email:     email,
		Filename:  filename.String(),
		CreatedAt: time.Now().UnixMilli(),
		FileURL:   objectUrl,
	}

	_, err = i.firestoreClient.Collection("images-test").Doc(filename.String()).Create(ctx, data)

	if err != nil {
		log.Printf("[ImageRepository.UploadImage] error write to firestore with error %v \n", err)
		return nil, err
	}

	return &data, nil
}

func (i *ImageRepository) GetDetectionResults(email string, filter *domain.PageFilter) ([]domain.Image, error) {
	result := []domain.Image{}

	ctx := context.Background()
	q := i.firestoreClient.Collection("images-test").Where("email", "==", email).OrderBy("createdAt", firestore.Desc)

	if filter.StartDate != 0 && filter.EndDate != 0 {
		q = q.Where("createdAt", ">=", filter.StartDate).Where("createdAt", "<=", filter.EndDate)
	}

	if len(filter.Labels) != 1 {
		q = q.Where("label", "in", filter.Labels)
	}

	if filter.After != "" {
		dsnap, err := i.firestoreClient.Collection("images-test").Doc(filter.After).Get(ctx)
		if err != nil {
			log.Printf("[ImageRepository.GetSingleDetection] error when retrieve dsnap with error %v \n", err)
			return nil, err
		}

		q = q.StartAfter(dsnap.Data()["createdAt"])
	}

	res := q.Limit(filter.PerPage).Documents(ctx)
	for {
		doc, err := res.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := domain.Image{
			Email:         fmt.Sprint(doc.Data()["email"]),
			Filename:      fmt.Sprint(doc.Data()["filename"]),
			FileURL:       fmt.Sprint(doc.Data()["fileURL"]),
			InferenceTime: doc.Data()["inferenceTime"].(int64),
			CreatedAt:     doc.Data()["createdAt"].(int64),
			DetectedAt:    doc.Data()["detectedAt"].(int64),
			IsDetected:    doc.Data()["isDetected"].(bool),
			Label:         fmt.Sprint(doc.Data()["label"]),
		}
		result = append(result, data)
	}
	return result, nil
}

func (i *ImageRepository) UpdateImageResult(payload domain.UpdateImagePayload) error {
	ctx := context.Background()

	// TODO: this query works when updating inferenceTime, detectedAt
	// and isDetected and still success when update label and .
	// However, upon successful updating label,
	// when querying the document, the result does not exists

	// TODO: figure out about indexing on label field
	// so later when read document, the query will give result
	_, err := i.firestoreClient.Collection("images-test").Doc(payload.Filename).Update(ctx, []firestore.Update{
		{
			Path:  "inferenceTime",
			Value: payload.InferenceTime,
		},
		{
			Path:  "detectedAt",
			Value: payload.DetectedAt,
		},
		{
			Path:  "isDetected",
			Value: true,
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

func (i *ImageRepository) GetSingleDetection(filename string) (*domain.Image, error) {
	ctx := context.Background()
	dsnap, err := i.firestoreClient.Collection("images-test").Doc(filename).Get(ctx)
	if err != nil {
		log.Printf("[ImageRepository.GetSingleDetection] error when querying to database with error %v \n", err)
		return nil, err
	}

	resp := domain.Image{
		Email:         fmt.Sprint(dsnap.Data()["email"]),
		Filename:      fmt.Sprint(dsnap.Data()["filename"]),
		FileURL:       fmt.Sprint(dsnap.Data()["fileURL"]),
		InferenceTime: dsnap.Data()["inferenceTime"].(int64),
		CreatedAt:     dsnap.Data()["createdAt"].(int64),
		DetectedAt:    dsnap.Data()["detectedAt"].(int64),
	}

	return &resp, nil
}
