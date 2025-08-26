package dto

type Response struct {
	Data  any    `json:"data"`
	Error string `json:"error,omitempty"`
}
