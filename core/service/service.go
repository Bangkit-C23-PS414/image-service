package service

import (
	"image-service/core/port"
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

func (i *ImageService) UploadImage(image *multipart.FileHeader) error {
	i.repo.UploadImage(image)
	return nil
}
