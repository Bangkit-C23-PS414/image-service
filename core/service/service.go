package service

import (
	"image-service/core/domain"
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

func (i *ImageService) UploadImage(email string, image *multipart.File) error {
	err := i.repo.UploadImage(email, image)
	if err != nil {
		log.Printf("[ImageService.UploadImage] error when uploading image with error %v \n", err)
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

func (i *ImageService) UpdateImageResult(payload domain.UpdateImagePayload) error {
	err := i.repo.UpdateImageResult(payload)
	if err != nil {
		log.Printf("[ImageService.UpdateImageResult] error update image result with error %v \n", err)
		return err
	}
	return nil
}

func (i *ImageService) GetSingleDetection(filename string) (*domain.Image, error) {
	res, err := i.repo.GetSingleDetection(filename)
	if err != nil {
		log.Printf("[ImageService.UpdateImageResult] error when retrieve data from database with error %v \n", err)
		return nil, err
	}
	return res, nil
}
