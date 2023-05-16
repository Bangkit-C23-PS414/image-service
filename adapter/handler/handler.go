package handler

import (
	"image-service/core/service"
	"log"
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
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 20*1024*1024)
	file, h, err := r.FormFile("image")
	if err != nil {
		log.Printf("[ImageHttpHandler.UploadImage] fail to read from file with error %v \n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fileContentType := h.Header.Get("Content-Type")
	if fileContentType != "image/jpeg" || fileContentType != "image/png" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	i.imageService.UploadImage(h)
}

func InitHttpServer(imageService service.ImageService) {
	mux := http.NewServeMux()
	imageHandler := NewImageHttpHandler(imageService)
	mux.HandleFunc("/image-detections", imageHandler.UploadImage)
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("error listening to port 8080 with error %v \n", err)
		return
	}
}
