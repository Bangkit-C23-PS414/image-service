package service

import (
	"context"
	"image"
	"image-service/core/domain"
	"image-service/core/port"
	"log"
	"mime/multipart"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/bbrks/go-blurhash"
)

type ImageService struct {
	repo         port.ImageRepository
	pubsubClient pubsub.Client
}

func NewImageService(repo port.ImageRepository) (*ImageService, error) {
	ctx := context.Background()
	projectId := os.Getenv("ML_SERVICE_PROJECT_ID")
	pubsubClient, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		log.Printf("failed to initialize pubsub client with error %v \n", err)
		return nil, err
	}
	return &ImageService{
		repo:         repo,
		pubsubClient: *pubsubClient,
	}, nil
}

func (i *ImageService) UploadImage(email string, image multipart.File) (*domain.UploadImageResponse, error) {
	res, err := i.repo.UploadImage(email, image)
	if err != nil {
		log.Printf("[ImageService.UploadImage] error when uploading image with error %v \n", err)
		return nil, err
	}
	// data := domain.SendToMLPayload{
	// 	Filename: res.Filename,
	// 	FileURL:  res.FileURL,
	// }
	// buf := new(bytes.Buffer)
	// err = json.NewEncoder(buf).Encode(&payload)
	// payload, err := json.Marshal(data)
	// if err != nil {
	// 	log.Printf("[ImageService.UploadImage] error when encode to json with error %v \n", err)
	// 	return nil, err
	// }

	// projectId := os.Getenv("CAPSTONE_PROJECT_ID")
	// topic := i.pubsubClient.TopicInProject("upload-image", projectId)

	// publishRes := topic.Publish(context.TODO(), &pubsub.Message{
	// 	Data: buf.Bytes(),
	// })

	// _, err = publishRes.Get(context.TODO())
	// if err != nil {
	// 	log.Printf("[ImageService.UploadImage] error when publish to pub/sub with error %v \n", err)
	// 	return nil, err
	// }

	// req, err := http.NewRequest(http.MethodPut, "https://c23-ps414-ml-service.et.r.appspot.com/predict", bytes.NewReader(payload))
	// if err != nil {
	// 	log.Printf("[ImageService.UploadImage] error creating request with error %v \n", err)
	// 	return nil, err
	// }
	// req.Header.Set("Content-Type", "application/json")
	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	log.Printf("[ImageService.UploadImage] error sending request to ML with error %v \n", err)
	// 	return nil, err
	// }
	// _, err = io.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Printf("[ImageService.UploadImage] error read response body with error %v \n", err)
	// 	return nil, err
	// }

	return res, nil
}

func (i *ImageService) UpdateBlurHash(filename string, file multipart.File) error {
	_, err := file.Seek(0, 0)
	if err != nil {
		log.Printf("[ImageRepository.UpdateBlurHash] error seeking with error %v \n", err)
		return err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("[ImageRepository.UpdateBlurHash] error decode image with error %v \n", err)
		return err
	}
	hash, err := blurhash.Encode(3, 3, img)
	if err != nil {
		log.Printf("[ImageRepository.UpdateBlurHash] error when encode file with error %v \n", err)
		return err
	}
	err = i.repo.UpdateBlurHash(filename, hash)
	if err != nil {
		log.Printf("[ImageRepository.UpdateBlurHash] error when performing blur hash update with error %v \n", err)
		return err
	}
	return nil
}

func (i *ImageService) GetDetectionResults(email string, filter *domain.PageFilter) ([]domain.Image, error) {
	res, err := i.repo.GetDetectionResults(email, filter)
	if err != nil {
		log.Printf("[ImageService.GetDetectionResults] error when retrieve detection results with error %v \n", err)
		return nil, err
	}
	return res, nil
}

func (i *ImageService) UpdateImageResult(payload domain.UpdateImagePayloadData) error {
	err := i.repo.UpdateImageResult(payload)
	if err != nil {
		log.Printf("[ImageService.UpdateImageResult] error update image result with error %v \n", err)
		return err
	}
	return nil
}

func (i *ImageService) GetSingleDetection(email, filename string) (*domain.Image, error) {
	res, err := i.repo.GetSingleDetection(email, filename)
	if err != nil {
		log.Printf("[ImageService.UpdateImageResult] error when retrieve data from database with error %v \n", err)
		return nil, err
	}
	return res, nil
}
