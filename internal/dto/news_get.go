package dto

import "time"

type NewsGetRequest struct {
	Page   int32    `json:"page"`
	Limit  int32    `json:"limit"`
	Source []string `json:"source,omitempty"`
}

type NewsGetResponse struct {
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`
	Link        string    `json:"link"`
}
