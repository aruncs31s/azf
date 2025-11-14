package api_usage

import "time"

// APIUsageLog represents a record of API endpoint usage
type APIUsageLog struct {
	ID             string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Endpoint       string    `gorm:"index;type:varchar(255)" json:"endpoint"`
	Method         string    `gorm:"index;type:varchar(10)" json:"method"`
	StatusCode     int       `gorm:"index" json:"status_code"`
	ResponseTime   int64     `gorm:"type:bigint" json:"response_time_ms"`
	RequestSize    int       `gorm:"type:int" json:"request_size"`
	ResponseSize   int       `gorm:"type:int" json:"response_size"`
	UserID         *string   `gorm:"index;type:varchar(36)" json:"user_id"`
	ClientIP       string    `gorm:"type:varchar(45)" json:"client_ip"`
	UserAgent      string    `gorm:"type:text" json:"user_agent"`
	ErrorMessage   *string   `gorm:"type:text" json:"error_message"`
	RequestedAt    time.Time `gorm:"index" json:"requested_at"`
	LastAccessedAt time.Time `gorm:"index" json:"last_accessed_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// APIUsageStats represents aggregated statistics for API endpoints
type APIUsageStats struct {
	ID              string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Endpoint        string    `gorm:"index;type:varchar(255)" json:"endpoint"`
	Method          string    `gorm:"type:varchar(10)" json:"method"`
	TotalRequests   int64     `json:"total_requests"`
	SuccessRequests int64     `json:"success_requests"`
	ErrorRequests   int64     `json:"error_requests"`
	AvgResponseTime int64     `json:"avg_response_time_ms"`
	MaxResponseTime int64     `json:"max_response_time_ms"`
	MinResponseTime int64     `json:"min_response_time_ms"`
	Last24Hours     int64     `json:"last_24_hours"`
	LastAccessedAt  time.Time `json:"last_accessed_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// APIEndpointRanking represents endpoint usage ranking
type APIEndpointRanking struct {
	Endpoint        string `json:"endpoint"`
	Method          string `json:"method"`
	TotalRequests   int64  `json:"total_requests"`
	SuccessRequests int64  `json:"success_requests"`
	ErrorRequests   int64  `json:"error_requests"`
	AvgResponseTime int64  `json:"avg_response_time_ms"`
	Last24Hours     int64  `json:"last_24_hours"`
	Rank            int    `json:"rank"`
}

// TableName specifies the table name for APIUsageLog
func (APIUsageLog) TableName() string {
	return "api_usage_logs"
}

// TableName specifies the table name for APIUsageStats
func (APIUsageStats) TableName() string {
	return "api_usage_stats"
}
