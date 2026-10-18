package dto

// ErrorResponse is the body of every error.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse acknowledges a request that has nothing to return.
type SuccessResponse struct {
	Success bool `json:"success"`
}

// StatusResponse acknowledges a request with a status word.
type StatusResponse struct {
	Status string `json:"status"`
}
