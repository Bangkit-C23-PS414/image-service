package domain

import "time"

type Image struct {
	Label         string    `json:"label,omitempty"`
	InferenceTime int       `json:"inferenceTime,omitempty"`
	UploadedAt    time.Time `firestore:"uploadedAt,omitempty"`
	DetectedAt    time.Time `json:"detectedAt,omitempty"`
}

type ServerResponse struct {
	Message string `json:"Message,omitempty"`
	Data    interface{}
}
