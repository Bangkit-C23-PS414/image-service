package port

import (
	"mime/multipart"
)

type ImageService interface {
	UploadImage(*multipart.File) error
}

type ImageRepository interface {
	UploadImage(*multipart.File) error
}
