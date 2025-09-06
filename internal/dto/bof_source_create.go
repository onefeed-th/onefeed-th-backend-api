package dto

type CreateSourceRequest struct {
	Name   string `json:"name"`
	Tags   string `json:"tags"`
	RSSURL string `json:"rssUrl"`
}

type CreateSourceResponse struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Tags   string `json:"tags"`
	RSSURL string `json:"rssUrl"`
}
