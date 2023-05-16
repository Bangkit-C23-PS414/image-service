package domain

type Image struct {
	Filename      string `json:"Date,omitempty"`
	Username      string `json:"Username,omitempty"`
	Label         string `json:"Label,omitempty"`
	InferenceTime int    `json:"InferenceTime,omitempty"`
}
