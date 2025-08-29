package dto

import "time"

type NewsListGetRequest struct {
	Page   int32    `json:"page"`
	Limit  int32    `json:"limit"`
	Source []string `json:"source,omitempty"`
}

type NewsListGetResponse struct {
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"publishedAt"`
	Image       string    `json:"image"`
	Link        string    `json:"link"`
}
