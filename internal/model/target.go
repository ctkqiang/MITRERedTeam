package model

type Target struct {
	ID       string            `json:"id"`
	Host     string            `json:"host"`
	Port     int               `json:"port,omitempty"`
	Scheme   string            `json:"scheme,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
