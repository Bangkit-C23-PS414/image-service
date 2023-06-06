package domain

type Image struct {
	Email         string  `firestore:"email,omitempty" json:"email,omitempty"`
	Filename      string  `firestore:"filename,omitempty" json:"filename,omitempty"`
	Label         string  `firestore:"label" json:"label"`
	InferenceTime int64   `firestore:"inferenceTime" json:"inferenceTime"`
	CreatedAt     int64   `firestore:"createdAt" json:"createdAt"`
	DetectedAt    int64   `firestore:"detectedAt" json:"detectedAt"`
	Confidence    float64 `firestore:"confidence" json:"confidence"`
	FileURL       string  `firestore:"fileURL" json:"fileURL"`
	BlurHash      string  `firestore:"blurHash" json:"blurHash"`
	IsDetected    bool    `firestore:"isDetected" json:"isDetected"`
}

type UpdateImagePayloadData struct {
	Filename      string  `json:"filename"`
	Label         string  `firestore:"label" json:"label"`
	InferenceTime float32 `firestore:"inferenceTime" json:"inferenceTime"`
	DetectedAt    float32 `firestore:"detectedAt" json:"detectedAt"`
	Confidence    float64 `firestore:"confidence" json:"confidence"`
	IsDetected    bool    `firestore:"isDetected"`
}
type UpdateImagePayload struct {
	Message string                 `json:"message"`
	Data    UpdateImagePayloadData `json:"data"`
}

type SendToMLPayload struct {
	Filename string `json:"filename"`
	FileURL  string `json:"url"`
}

type ServerResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data"`
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
