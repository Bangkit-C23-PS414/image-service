package service

import "image-service/core/port"

type ImageService struct {
	repo port.ImageRepository
}

func NewImageService(repo port.ImageRepository) *ImageService {
	return &ImageService{
		repo: repo,
	}
}
