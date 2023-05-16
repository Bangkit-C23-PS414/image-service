package port

import "image-service/core/domain"

type ImageService interface {
	UploadImage(domain.Image) error
}

type ImageRepository interface {
	UploadImage(domain.Image) error
}
