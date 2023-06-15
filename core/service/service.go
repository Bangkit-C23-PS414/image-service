package service

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image-service/core/domain"
	"image-service/core/port"
	"log"
	"mime/multipart"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/bbrks/go-blurhash"
	"google.golang.org/api/option"
)

type ImageService struct {
	repo         port.ImageRepository
	pubsubClient pubsub.Client
}

func sendToPubsub(i *ImageService, payload domain.SendToMLPayload) error {
	data := domain.SendToMLPayload{
		Filename: payload.Filename,
		FileURL:  payload.FileURL,
	}
	var msg []byte
	// avroSource, err := os.ReadFile("ml-payload.avsc")
	// if err != nil {
	// 	log.Printf("[ImageService.UploadImage] read avrosource with error %v \n", err)
	// 	return nil, err
	// }
	// codec, err := goavro.NewCodec(string(avroSource))
	// if err != nil {
	// 	log.Printf("[ImageService.UploadImage] fail to use codec from avroSource with error %v \n", err)
	// 	return nil, err
	// }

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(&data)
	// payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("[ImageService.UploadImage] error when encode to json with error %v \n", err)
		return err
	}

	log.Println(string(buf.String()))
	topic := i.pubsubClient.Topic("upload-image")

	p := map[string]interface{}{
		"url":      payload.FileURL,
		"filename": payload.Filename,
	}
	log.Println(p)
	// cfg, err := topic.Config(context.TODO())
	// cfg.SchemaSettings.Encoding = pubsub.EncodingJSON
	// msg, err = codec.TextualFromNative(nil, payload)
	// if err != nil {
	// 	log.Printf("[ImageService.UploadImage] error when encode to json with error %v \n", err)
	// 	return nil, err
	// }
	// encoding := cfg.SchemaSettings.Encoding
	// switch encoding {
	// case pubsub.EncodingJSON:
	// 	msg, err = codec.TextualFromNative(nil, payload)
	// 	if err != nil {
	// 		log.Printf("[ImageService.UploadImage] error when encode to json with error %v \n", err)
	// 		return nil, err
	// 	}
	// default:
	// 	log.Printf("[ImageService.UploadImage] invalid encoding with error %v \n", err)
	// 	return nil, err
	// }
	log.Println(buf.String())
	publishRes := topic.Publish(context.TODO(), &pubsub.Message{
		Data: buf.Bytes(),
	})

	_, err = publishRes.Get(context.TODO())
	if err != nil {
		log.Printf("[ImageService.UploadImage] error when publish to pub/sub with error %v \n", err)
		return err
	}
	log.Printf("avro msg %v: \n", string(msg))
	return nil
}

func NewImageService(repo port.ImageRepository) (*ImageService, error) {
	ctx := context.Background()
	opt := option.WithCredentialsFile("pubsub-sa-key.json")
	projectId := os.Getenv("CAPSTONE_PROJECT_ID")
	pubsubClient, err := pubsub.NewClient(ctx, projectId, opt)
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
