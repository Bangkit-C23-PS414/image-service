package port

import (
	"mime/multipart"
)

type ImageService interface {
	UploadImage(string, *multipart.File) error
}

type ImageRepository interface {
	UploadImage(string, *multipart.File) error
}
