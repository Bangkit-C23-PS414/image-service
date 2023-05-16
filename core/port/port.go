package port

import (
	"mime/multipart"
)

type ImageService interface {
	UploadImage(*multipart.FileHeader) error
}

type ImageRepository interface {
	UploadImage(*multipart.FileHeader) error
}
