package dto

type GetAllSourceByPaginationRequest struct {
	PageLimit  int32 `json:"pageLimit"`
	PageOffset int32 `json:"pageOffset"`
}

type GetAllSourceByPaginationResponse struct {
	Sources []Source `json:"sources"`
}

type Source struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Tags   string `json:"tags"`
	RSSURL string `json:"rssUrl"`
}
