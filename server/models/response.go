package models

// Response represents a standard API response
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ResponseUser represents a user-specific API response
type ResponseUser struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
