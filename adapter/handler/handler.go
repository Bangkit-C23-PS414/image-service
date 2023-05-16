package handler

import (
	"image-service/core/service"
	"net/http"
)

type ImageHttpHandler struct {
	imageService service.ImageService
}

func NewImageHttpHandler(imageService service.ImageService) *ImageHttpHandler {
	return &ImageHttpHandler{
		imageService: imageService,
	}
}

func (i *ImageHttpHandler) UploadImage(w http.ResponseWriter, r *http.Request) {

}
