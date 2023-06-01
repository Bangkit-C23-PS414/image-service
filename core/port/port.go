package port

import (
	"image-service/core/domain"
	"mime/multipart"
)

type ImageService interface {
	UploadImage(string, *multipart.File) error
	GetDetectionResults(string, *domain.PageFilter) ([]domain.Image, error)
	UpdateImageResult(domain.UpdateImagePayload) error
	GetSingleDetection(string) (*domain.Image, error)
}

type ImageRepository interface {
	UploadImage(string, *multipart.File) error
	GetDetectionResults(string, *domain.PageFilter) ([]domain.Image, error)
	UpdateImageResult(domain.UpdateImagePayload) error
	GetSingleDetection(string) (*domain.Image, error)
}
