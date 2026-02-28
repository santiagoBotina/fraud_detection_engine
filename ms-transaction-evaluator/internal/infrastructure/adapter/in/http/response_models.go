package http

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Message string      `json:"message" example:"Transaction validation successful"`
	Data    interface{} `json:"data"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error   string `json:"error" example:"Validation failed"`
	Details string `json:"details" example:"customer email is invalid"`
}
