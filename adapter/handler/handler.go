package handler

import (
	"encoding/json"
	"image-service/core/domain"
	"image-service/core/service"
	"log"
	"net/http"
)

type ImageHttpHandler struct {
	imageService service.ImageService
}

func httpWriteResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
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
		httpWriteResponse(w, &domain.ServerResponse{
			Message: `Form data should be "image"`,
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fileContentType := h.Header.Get("Content-Type")
	if fileContentType != "image/jpg" && fileContentType != "image/jpeg" {
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Content-Type must be image/jmg or image/jpeg",
		})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "username should be filled",
		})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = i.imageService.UploadImage(username, &file)
	if err != nil {
		log.Printf("[ImageHttpHandler.UploadImage] error when uploading image with error %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Error upload image to database",
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func InitHttpServer(imageService service.ImageService) {
	mux := http.NewServeMux()
	imageHandler := NewImageHttpHandler(imageService)
	mux.HandleFunc("/image-detections/create", imageHandler.UploadImage)
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
