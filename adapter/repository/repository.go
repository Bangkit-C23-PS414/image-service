package repository

import (
	"context"
	"fmt"
	"image"
	"image-service/core/domain"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/bbrks/go-blurhash"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type ImageRepository struct {
	firestoreClient firestore.Client
	gcsClient       storage.Client
}

func generateSignedURL(i *ImageRepository, filename string) (string, error) {
	bktName := os.Getenv("CAPSTONE_IMAGE_BUCKET")
	gcsOpt := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(7 * 24 * time.Hour),
	}
	objectUrl, err := i.gcsClient.Bucket(bktName).SignedURL(fmt.Sprintf("images/%v", filename), gcsOpt)
	if err != nil {
		return "", err
	}
	return objectUrl, nil
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

func (i *ImageRepository) UploadImage(email string, file multipart.File) (*domain.Image, error) {
	ctx := context.Background()
	filename := uuid.New()
	bktName := os.Getenv("CAPSTONE_IMAGE_BUCKET")
	w := i.gcsClient.Bucket(bktName).Object("images/" + filename.String()).NewWriter(ctx)
	_, err := io.Copy(w, file)
	if err != nil {
		log.Printf("[ImageRepository.UploadImage] error writing to gcs bucket with error %v \n", err)
		return nil, err
	}
	if err = w.Close(); err != nil {
		log.Printf("[ImageRepository.UploadImage] error closing file with error %v \n", err)
		return nil, err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		log.Printf("[ImageRepository.UploadImage] error seeking with error %v \n", err)
		return nil, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("[ImageRepository.UploadImage] error decode image with error %v \n", err)
		return nil, err
	}
	hash, err := blurhash.Encode(3, 3, img)
	if err != nil {
		log.Printf("[ImageRepository.UploadImage] error when performing blur hash with error %v \n", err)
		return nil, err
	}

	objectUrl, err := generateSignedURL(i, filename.String())
	if err != nil {
		log.Printf("[ImageRepository.UploadImage] error when generate objectURl with error %v \n", err)
		return nil, err
	}

	data := domain.Image{
		Email:     email,
		Filename:  filename.String(),
		CreatedAt: time.Now().UnixMilli(),
		FileURL:   objectUrl,
		BlurHash:  hash,
	}

	_, err = i.firestoreClient.Collection("images").Doc(filename.String()).Create(ctx, data)

	if err != nil {
		log.Printf("[ImageRepository.UploadImage] error write to firestore with error %v \n", err)
		return nil, err
	}

	return &data, nil
}

func (i *ImageRepository) GetDetectionResults(email string, filter *domain.PageFilter) ([]domain.Image, error) {
	result := []domain.Image{}

	ctx := context.Background()
	q := i.firestoreClient.Collection("images").Where("email", "==", email).OrderBy("createdAt", firestore.Desc)

	if filter.StartDate != 0 && filter.EndDate != 0 {
		q = q.Where("createdAt", ">=", filter.StartDate).Where("createdAt", "<=", filter.EndDate)
	}

	if len(filter.Labels) > 0 {
		filter.Labels = append(filter.Labels, "")
		log.Println("execute label filter")
		q = q.Where("label", "in", filter.Labels)
	}

	if filter.After != "" {
		dsnap, err := i.firestoreClient.Collection("images").Doc(filter.After).Get(ctx)
		if err != nil {
			log.Printf("[ImageRepository.GetDetectionResults] error when retrieve dsnap with error %v \n", err)
			return nil, err
		}
		log.Printf("filename: %v \n", dsnap.Data()["filename"])

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

		filename := fmt.Sprint(doc.Data()["filename"])
		objectURL, err := generateSignedURL(i, filename)
		if err != nil {
			log.Printf("[ImageRepository.GetDetectionResults] error when generate objectURL with error %v \n", err)
			return nil, err
		}

		data := domain.Image{
			Email:         fmt.Sprint(doc.Data()["email"]),
			Filename:      filename,
			FileURL:       objectURL,
			InferenceTime: doc.Data()["inferenceTime"].(int64),
			CreatedAt:     doc.Data()["createdAt"].(int64),
			DetectedAt:    doc.Data()["detectedAt"].(int64),
			Confidence:    doc.Data()["confidence"].(float64),
			IsDetected:    doc.Data()["isDetected"].(bool),
			Label:         fmt.Sprint(doc.Data()["label"]),
			BlurHash:      fmt.Sprint(doc.Data()["blurHash"]),
		}
		result = append(result, data)
	}
	return result, nil
}

func (i *ImageRepository) UpdateImageResult(payload domain.UpdateImagePayloadData) error {
	ctx := context.Background()
	_, err := i.firestoreClient.Collection("images").Doc(payload.Filename).Update(ctx, []firestore.Update{
		{
			Path:  "inferenceTime",
			Value: int64(payload.InferenceTime),
		},
		{
			Path:  "detectedAt",
			Value: int64(payload.DetectedAt),
		},
		{
			Path:  "isDetected",
			Value: true,
		},
		{
			Path:  "label",
			Value: payload.Label,
		},
		{
			Path:  "confidence",
			Value: float64(payload.Confidence),
		},
	})

	if err != nil {
		log.Printf("[ImageRepository.UpdateImageResult] error when update image result with error %v", err)
		return err
	}
	return nil
}

func (i *ImageRepository) GetSingleDetection(email, filename string) (*domain.Image, error) {
	ctx := context.Background()
	docs := i.firestoreClient.Collection("images").Where("email", "==", email).Where("filename", "==", filename).Documents(ctx)

	var resp domain.Image
	for {
		doc, err := docs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		filename := fmt.Sprint(doc.Data()["filename"])
		objectURL, err := generateSignedURL(i, filename)
		if err != nil {
			log.Printf("[ImageRepository.GetSingleDetection] error when generate objectURL with error %v \n", err)
			return nil, err
		}

		resp = domain.Image{
			Email:         fmt.Sprint(doc.Data()["email"]),
			Filename:      filename,
			FileURL:       objectURL,
			InferenceTime: doc.Data()["inferenceTime"].(int64),
			CreatedAt:     doc.Data()["createdAt"].(int64),
			DetectedAt:    doc.Data()["detectedAt"].(int64),
			Confidence:    doc.Data()["confidence"].(float64),
			IsDetected:    doc.Data()["isDetected"].(bool),
			Label:         fmt.Sprint(doc.Data()["label"]),
			BlurHash:      fmt.Sprint(doc.Data()["blurHash"]),
		}
	}

	return &resp, nil
}
