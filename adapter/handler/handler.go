package handler

import (
	"encoding/json"
	"fmt"
	"image-service/core/domain"
	"image-service/core/service"
	"image-service/core/util"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/api/iterator"
)

var JWT_SIGNATURE_KEY = []byte(os.Getenv("JWT_SIGNATURE_KEY"))

type ImageHttpHandler struct {
	imageService service.ImageService
}

func checkToken(w http.ResponseWriter, r *http.Request) (jwt.MapClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.Contains(authHeader, "Bearer") {
		return nil, fmt.Errorf("invalid signing method")
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
		return nil, fmt.Errorf(err.Error())
	}

	claim, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Printf("[Server.tokenHandler] token is invalid %v \n", err)
		return nil, fmt.Errorf(err.Error())
	}

	return claim, nil
}

func httpWriteResponse(w http.ResponseWriter, response interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
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
			Message: "error read image",
		}, http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fileContentType := h.Header.Get("Content-Type")
	if fileContentType != "image/jpg" && fileContentType != "image/jpeg" {
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Content-Type must be image/jmg or image/jpeg",
		}, http.StatusBadRequest)
		return
	}

	claim, err := checkToken(w, r)
	if err != nil {
		log.Printf("unable to retrieve token claim with error %v \n", err)
		httpWriteResponse(w, domain.ServerResponse{
			Message: "error retrieve claim",
		}, http.StatusInternalServerError)
		return
	}

	email := fmt.Sprint(claim["email"])
	if email == "" {
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "email should be filled",
		}, http.StatusBadRequest)
		return
	}

	res, err := i.imageService.UploadImage(email, file)
	if err != nil {
		log.Printf("[ImageHttpHandler.UploadImage] error when uploading image with error %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Error upload image to database",
		}, http.StatusInternalServerError)
		return
	}

	log.Printf("[ImageHttpHandler.UploadImage] [/image-detections/create] success upload image to database from payload: %v \n", res)
	httpWriteResponse(w, &domain.ServerResponse{
		Message: "Success",
		Data:    res,
	}, http.StatusAccepted)
}

func (i *ImageHttpHandler) GetDetectionResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	claim, err := checkToken(w, r)
	if err != nil {
		log.Printf("[ImageHttpHandler.GetDetectionResults] error when checking token with error %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: err.Error(),
		}, http.StatusUnauthorized)
		return
	}

	filter := util.PageFilter(r)
	email := fmt.Sprint(claim["email"])
	res, err := i.imageService.GetDetectionResults(email, &filter)
	if err == iterator.Done {
		log.Println("[ImageHttpHanndler.GetSingleDetection]unable to iterate next document, the cursor reached the end of documents")
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "error when read document, the cursor has reach the limit",
		}, http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("[ImageHttpHandler.GetDetectionResults] error when retrieve detection results with error %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Error read image from database",
		}, http.StatusInternalServerError)
		return
	}

	log.Printf("[ImageHttpHandler.GetDetectionResults] [/image-detections/fetch] success retrieve documents with values: %v \n", res)
	httpWriteResponse(w, domain.ServerResponse{
		Message: "Success",
		Data:    res,
	}, http.StatusOK)
}

func (i *ImageHttpHandler) UpdateImageResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload domain.UpdateImagePayloadData
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ImageHttpHandler.UpdateImageResult] error read request body with error %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Error read request body",
		}, http.StatusInternalServerError)
		return
	}

	data, err := url.ParseQuery(string(body))
	if err != nil {
		log.Printf("[ImageHttpHandler.UpdateImageResult] error parsequery  %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Error parse form",
		}, http.StatusInternalServerError)
		return
	}

	filename := data.Get("filename")
	confidence := data.Get("confidence")
	detectedAt := data.Get("detectedAt")
	inferenceTime := data.Get("inferenceTime")
	label := data.Get("label")

	fConfidence, err := strconv.ParseFloat(confidence, 64)
	if err != nil {
		log.Printf("[ImageHttpHandler.UpdateImageResult] error when convert confident from string to float64 with error %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Error update result to database",
		}, http.StatusInternalServerError)
		return
	}

	fDetectedAt, err := strconv.ParseFloat(detectedAt, 32)
	if err != nil {
		log.Printf("[ImageHttpHandler.UpdateImageResult] error when convert detected from string to float32 with error %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Error update result to database",
		}, http.StatusInternalServerError)
		return
	}

	fInferenceTime, err := strconv.ParseFloat(inferenceTime, 32)
	if err != nil {
		log.Printf("[ImageHttpHandler.UpdateImageResult] error when convert inferenceTime from string to float32 with error %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Error update result to database",
		}, http.StatusInternalServerError)
		return
	}

	if filename == "" {
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "filename should be filled",
		}, http.StatusBadRequest)
		return
	}

	if label == "" {
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "label should be filled",
		}, http.StatusBadRequest)
		return
	}

	if fInferenceTime == 0 {
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "inferenceTime should be filled",
		}, http.StatusBadRequest)
		return
	}

	if fDetectedAt == 0 {
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "detectedAt should be filled",
		}, http.StatusBadRequest)
		return
	}

	if fConfidence == 0 {
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "detectedAt should be filled",
		}, http.StatusBadRequest)
		return
	}

	payload.Filename = filename
	payload.Label = label
	payload.DetectedAt = float32(fDetectedAt)
	payload.InferenceTime = float32(fInferenceTime)
	payload.Confidence = fConfidence

	err = i.imageService.UpdateImageResult(payload)
	if err != nil {
		log.Printf("[ImageHttpHandler.UpdateImageResult] error when update detection with error %v \n", err)
		httpWriteResponse(w, &domain.ServerResponse{
			Message: "Error update result to database",
		}, http.StatusInternalServerError)
		return
	}

	log.Printf("[ImageHttpHandler.UpdateImageResult] [/image-detections/update] success upload detection data from payload: %v \n", payload)
	httpWriteResponse(w, domain.ServerResponse{
		Message: "Success",
	}, http.StatusOK)
}

func (i *ImageHttpHandler) GetSingleDetection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpWriteResponse(w, domain.ServerResponse{
			Message: "invalid method",
		}, http.StatusMethodNotAllowed)
		return
	}

	claim, err := checkToken(w, r)
	if err != nil {
		log.Printf("[ImageHttpHandler.GetSingleDetection] error when checking token with error %v \n", err.Error())
		httpWriteResponse(w, domain.ServerResponse{
			Message: "error checking token",
		}, http.StatusInternalServerError)
		return
	}

	email := fmt.Sprint(claim["email"])

	path := strings.Split(r.URL.Path, "/image-detections/fetch/")

	res, err := i.imageService.GetSingleDetection(email, path[1])
	if err != nil {
		log.Printf("[ImageHttpHandler.GetSingleDetection] error when retireve data from database with error %v \n", err.Error())
		httpWriteResponse(w, domain.ServerResponse{
			Message: "error retrieve data from database",
		}, http.StatusInternalServerError)
		return
	}

	if res.Filename == "" {
		httpWriteResponse(w, domain.ServerResponse{
			Message: "data not found",
		}, http.StatusNotFound)
		return
	}

	log.Printf("[ImageHttpHandler.GetSingleDetection] [/image-detections/fetch/] success retrieve single document with response: %v \n", res)
	httpWriteResponse(w, domain.ServerResponse{
		Message: "success",
		Data:    res,
	}, http.StatusOK)
}

func InitHttpServer(imageService service.ImageService) {
	mux := http.NewServeMux()
	imageHandler := NewImageHttpHandler(imageService)
	mux.HandleFunc("/image-detections/create", imageHandler.UploadImage)
	mux.HandleFunc("/image-detections/fetch", imageHandler.GetDetectionResults)
	mux.HandleFunc("/image-detections/update", imageHandler.UpdateImageResult)
	mux.HandleFunc("/image-detections/fetch/", imageHandler.GetSingleDetection)
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Println("serving at port 8080")
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("error listening to port 8080 with error %v \n", err)
		return
	}
}
