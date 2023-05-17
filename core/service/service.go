package service

import (
	"image-service/core/port"
	"log"
	"mime/multipart"
)

type ImageService struct {
	repo port.ImageRepository
}

func NewImageService(repo port.ImageRepository) *ImageService {
	return &ImageService{
		repo: repo,
	}
}

func (i *ImageService) UploadImage(image *multipart.File) error {
	err := i.repo.UploadImage(image)
	if err != nil {
		log.Printf("[ImageService.UploadImage] error when uploading image with error %v \n", err)
		return err
	}
	return nil
}
