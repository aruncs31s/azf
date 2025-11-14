package dto

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=1,max=100" form:"username"`
	Password string `json:"password" binding:"required,min=6,max=255" form:"password"`
	// CollegeID string `json:"college_id" binding:"required"`
}

type AdminLoginResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	SessionID string    `json:"session_id,omitempty"`
	Admin     AdminInfo `json:"admin,omitempty"`
	JWT       string    `json:"jwt,omitempty"`
	Error     string    `json:"error,omitempty"`
	Timestamp string    `json:"timestamp"`
}

// AdminInfo represents basic admin information
type AdminInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
