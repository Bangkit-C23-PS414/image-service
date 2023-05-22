package port

import (
	"image-service/core/domain"
	"mime/multipart"
)

type ImageService interface {
	UploadImage(string, *multipart.File) error
	GetDetectionResults(string) ([]domain.Image, error)
}

type ImageRepository interface {
	UploadImage(string, *multipart.File) error
	GetDetectionResults(string) ([]domain.Image, error)
}
