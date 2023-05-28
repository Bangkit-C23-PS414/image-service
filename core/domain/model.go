package domain

import "time"

type Image struct {
	Email      string    `firestore:"email,omitempty"`
	Filename      string    `firestore:"filename,omitempty"`
	Label         string    `firestore:"label"`
	InferenceTime int64     `firestore:"inferenceTime"`
	UploadedAt    time.Time `firestore:"uploadedAt"`
	DetectedAt    time.Time `firestore:"detectedAt"`
}

type UpdateImagePayload struct {
	Filename      string `firestore:"filename,omitempty"`
	Label         string `firestore:"label"`
	InferenceTime int64  `firestore:"inferenceTime"`
}

type ServerResponse struct {
	Message string `json:"Message,omitempty"`
	Data    interface{}
}

type UserData struct {
	Username string `json:"Message,omitempty"`
}
