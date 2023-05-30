package handler

import (
	"encoding/json"
	"fmt"
	"image-service/core/domain"
	"image-service/core/service"
	"image-service/core/util"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

var JWT_SIGNATURE_KEY = []byte(os.Getenv("JWT_SIGNATURE_KEY"))

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
		w.WriteHeader(http.StatusInternalServerError)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "error read image",
		})
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

	email := r.FormValue("email")
	if email == "" {
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "email should be filled",
		})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = i.imageService.UploadImage(email, &file)
	if err != nil {
		log.Printf("[ImageHttpHandler.UploadImage] error when uploading image with error %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Error upload image to database",
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	httpWriteResponse(w, &domain.ServerResponse{
		Message: "Success",
	})
}

func (i *ImageHttpHandler) GetDetectionResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	authHeader := r.Header.Get("Authorization")
	if !strings.Contains(authHeader, "Bearer") {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", -1)
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if method, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		} else if method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("invalid signing method")
		}
		return JWT_SIGNATURE_KEY, nil
	})

	if err != nil {
		log.Printf("[Server.tokenHandler] unable to parse token with error %v \n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	claim, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Printf("[Server.tokenHandler] token is invalid %v \n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	filter := util.PageFilter(r)
	email := fmt.Sprint(claim["email"])
	res, err := i.imageService.GetDetectionResults(email, &filter)
	if err != nil {
		log.Printf("[ImageHttpHandler.GetDetectionResults] error when retrieve detection results with error %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Error read image from database",
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	httpWriteResponse(w, domain.ServerResponse{
		Message: "Success",
		Data:    res,
	})
}

func (i *ImageHttpHandler) UpdateImageResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	filename := r.FormValue("filename")
	if filename == "" {
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "filename should be filled",
		})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	label := r.FormValue("label")
	if label == "" {
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "label should be filled",
		})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	inferenceTime := r.FormValue("inferenceTime")
	if inferenceTime == "" {
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "inferenceTime should be filled",
		})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	intInferenceTime, err := strconv.ParseInt(inferenceTime, 10, 64)
	if err != nil {
		log.Printf("[ImageHttpHandler.UpdateImageResult] error parsing inference time to int64 with error %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Error parsing inference time",
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	payload := domain.UpdateImagePayload{
		Filename:      filename,
		Label:         label,
		InferenceTime: intInferenceTime,
	}
	err = i.imageService.UpdateImageResult(payload)
	if err != nil {
		log.Printf("[ImageHttpHandler.UpdateImageResult] error when update detection with error %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Error update result to database",
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	httpWriteResponse(w, domain.ServerResponse{
		Message: "Success",
	})
}

func InitHttpServer(imageService service.ImageService) {
	mux := http.NewServeMux()
	imageHandler := NewImageHttpHandler(imageService)
	mux.HandleFunc("/image-detections/create", imageHandler.UploadImage)
	mux.HandleFunc("/image-detections/fetch", imageHandler.GetDetectionResults)
	mux.HandleFunc("/image-detections/update", imageHandler.UpdateImageResult)
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
