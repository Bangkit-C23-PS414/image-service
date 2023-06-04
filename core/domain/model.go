package domain

type Image struct {
	Email         string `firestore:"email,omitempty" json:"email,omitempty"`
	Filename      string `firestore:"filename,omitempty" json:"filename,omitempty"`
	Label         string `firestore:"label" json:"label"`
	InferenceTime int64  `firestore:"inferenceTime" json:"inferenceTime"`
	CreatedAt     int64  `firestore:"createdAt" json:"createdAt"`
	DetectedAt    int64  `firestore:"detectedAt" json:"detectedAt"`
	Confidence    int64  `firestore:"confidence" json:"confidence"`
	FileURL       string `firestore:"fileURL" json:"fileURL"`
	IsDetected    bool   `firestore:"isDetected" json:"isDetected"`
}

type UpdateImagePayload struct {
	Filename      string
	Label         string `firestore:"label"`
	InferenceTime int64  `firestore:"inferenceTime"`
	DetectedAt    int64  `firestore:"detectedAt"`
	Confidence    int64  `firestore:"confidence"`
	IsDetected    bool   `firestore:"isDetected"`
}

type ServerResponse struct {
	Message string `json:"Message,omitempty"`
	Data    interface{}
}

type UserData struct {
	Username string `json:"Message,omitempty"`
}

type PageFilter struct {
	Page      int      `json:"page"`
	PerPage   int      `json:"perPage"`
	StartDate int      `json:"startDate"`
	EndDate   int      `json:"endDate"`
	Labels    []string `json:"labels"`
	After     string   `json:"after"`
}
