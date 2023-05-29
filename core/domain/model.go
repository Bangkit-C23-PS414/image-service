package domain

import "time"

type Image struct {
	Email         string    `firestore:"email,omitempty" json:"email,omitempty"`
	Filename      string    `firestore:"filename,omitempty" json:"filename,omitempty"`
	Label         string    `firestore:"label" json:"label"`
	InferenceTime int64     `firestore:"inferenceTime" json:"inferenceTime"`
	UploadedAt    time.Time `firestore:"uploadedAt" json:"uploadedAt"`
	DetectedAt    time.Time `firestore:"detectedAt" json:"detectedAt"`
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

type PageFilter struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page`
}
