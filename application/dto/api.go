package dto

type ResponseMeta struct {
	APIVersion        string `json:"api_version"`
	AuthorizationMode string `json:"authorization_mode"`
	ResponseTimeMs    string `json:"response_time_ms"`
	RequestID         string `json:"request_id"`
}
