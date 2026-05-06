package handler

// ErrorResponse is the standard error envelope returned by all endpoints.
type ErrorResponse struct {
	Error string `json:"error"`
}
