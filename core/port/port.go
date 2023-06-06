package port

import (
	"image-service/core/domain"
	"mime/multipart"
)

type ImageService interface {
	UploadImage(string, multipart.File) (*domain.Image, error)
	GetDetectionResults(string, *domain.PageFilter) ([]domain.Image, error)
	UpdateImageResult(domain.UpdateImagePayloadData) error
	GetSingleDetection(string, string) (*domain.Image, error)
}

type ImageRepository interface {
	UploadImage(string, multipart.File) (*domain.Image, error)
	GetDetectionResults(string, *domain.PageFilter) ([]domain.Image, error)
	UpdateImageResult(domain.UpdateImagePayloadData) error
	GetSingleDetection(string, string) (*domain.Image, error)
}
