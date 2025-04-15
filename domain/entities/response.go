package entities

type ErrResponse struct {
	Message        string `json:"message"`
	HTTPStatusCode int    `json:"-"`
}
