package model

import (
	"testing"
	"time"
)

// TestAuditActionCreation tests AuditAction creation and validation
func TestAuditActionCreation(t *testing.T) {
	tests := []struct {
		name        string
		action      string
		expectError bool
	}{
		{
			name:        "Valid CREATE action",
			action:      "CREATE",
			expectError: false,
		},
		{
			name:        "Valid UPDATE action",
			action:      "UPDATE",
			expectError: false,
		},
		{
			name:        "Valid DELETE action",
			action:      "DELETE",
			expectError: false,
		},
		{
			name:        "Valid READ action",
			action:      "READ",
			expectError: false,
		},
		{
			name:        "Empty action",
			action:      "",
			expectError: true,
		},
		{
			name:        "Invalid action",
			action:      "INVALID_ACTION",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := NewAuditAction(tt.action)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for action %s, but got nil", tt.action)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for valid action %s: %v", tt.action, err)
				}
				if action == nil {
					t.Errorf("Expected non-nil action for %s", tt.action)
				} else {
					if action.Value() != tt.action {
						t.Errorf("Expected action value %s, got %s", tt.action, action.Value())
					}
				}
			}
		})
	}
}

// TestAuditActionEquality tests AuditAction equality
func TestAuditActionEquality(t *testing.T) {
	action1, _ := NewAuditAction("CREATE")
	action2, _ := NewAuditAction("CREATE")
	action3, _ := NewAuditAction("UPDATE")

	tests := []struct {
		name     string
		a1       *AuditAction
		a2       *AuditAction
		expected bool
	}{
		{
			name:     "Same action values",
			a1:       action1,
			a2:       action2,
			expected: true,
		},
		{
			name:     "Different action values",
			a1:       action1,
			a2:       action3,
			expected: false,
		},
		{
			name:     "Both nil",
			a1:       nil,
			a2:       nil,
			expected: true,
		},
		{
			name:     "One nil",
			a1:       action1,
			a2:       nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a1.Equals(tt.a2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestAdminIDCreation tests AdminID creation and validation
func TestAdminIDCreation(t *testing.T) {
	tests := []struct {
		name        string
		adminID     string
		expectError bool
	}{
		{
			name:        "Valid admin ID",
			adminID:     "admin123",
			expectError: false,
		},
		{
			name:        "Valid UUID admin ID",
			adminID:     "550e8400-e29b-41d4-a716-446655440000",
			expectError: false,
		},
		{
			name:        "Empty admin ID",
			adminID:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := NewAdminID(tt.adminID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for admin ID %s", tt.adminID)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for valid admin ID %s: %v", tt.adminID, err)
				}
				if id == nil {
					t.Errorf("Expected non-nil admin ID for %s", tt.adminID)
				} else {
					if id.Value() != tt.adminID {
						t.Errorf("Expected admin ID value %s, got %s", tt.adminID, id.Value())
					}
				}
			}
		})
	}
}

// TestAdminIDEquality tests AdminID equality
func TestAdminIDEquality(t *testing.T) {
	id1, _ := NewAdminID("admin123")
	id2, _ := NewAdminID("admin123")
	id3, _ := NewAdminID("admin456")

	tests := []struct {
		name     string
		id1      *AdminID
		id2      *AdminID
		expected bool
	}{
		{
			name:     "Same admin IDs",
			id1:      id1,
			id2:      id2,
			expected: true,
		},
		{
			name:     "Different admin IDs",
			id1:      id1,
			id2:      id3,
			expected: false,
		},
		{
			name:     "Both nil",
			id1:      nil,
			id2:      nil,
			expected: true,
		},
		{
			name:     "One nil",
			id1:      id1,
			id2:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.id1.Equals(tt.id2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestLogStatusCreation tests LogStatus creation
func TestLogStatusCreation(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		expectError bool
	}{
		{
			name:        "Valid SUCCESS status",
			status:      "SUCCESS",
			expectError: false,
		},
		{
			name:        "Valid FAILURE status",
			status:      "FAILURE",
			expectError: false,
		},
		{
			name:        "Valid PENDING status",
			status:      "PENDING",
			expectError: false,
		},
		{
			name:        "Empty status",
			status:      "",
			expectError: true,
		},
		{
			name:        "Invalid status",
			status:      "INVALID",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := NewLogStatus(tt.status)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for status %s", tt.status)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for valid status %s: %v", tt.status, err)
				}
				if status == nil {
					t.Errorf("Expected non-nil status for %s", tt.status)
				} else {
					if status.Value() != tt.status {
						t.Errorf("Expected status value %s, got %s", tt.status, status.Value())
					}
				}
			}
		})
	}
}

// TestLogStatusEquality tests LogStatus equality
func TestLogStatusEquality(t *testing.T) {
	status1, _ := NewLogStatus("SUCCESS")
	status2, _ := NewLogStatus("SUCCESS")
	status3, _ := NewLogStatus("FAILED")

	tests := []struct {
		name     string
		s1       *LogStatus
		s2       *LogStatus
		expected bool
	}{
		{
			name:     "Same status values",
			s1:       status1,
			s2:       status2,
			expected: true,
		},
		{
			name:     "Different status values",
			s1:       status1,
			s2:       status3,
			expected: false,
		},
		{
			name:     "Both nil",
			s1:       nil,
			s2:       nil,
			expected: true,
		},
		{
			name:     "One nil",
			s1:       status1,
			s2:       nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.s1.Equals(tt.s2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestAuditLogTimestamp tests that audit logs can track timestamps
func TestAuditLogTimestamp(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		timestamp time.Time
		validate  func(time.Time, time.Time) bool
	}{
		{
			name:      "Current timestamp",
			timestamp: now,
			validate: func(recorded, expected time.Time) bool {
				return recorded.Unix() == expected.Unix()
			},
		},
		{
			name:      "Past timestamp",
			timestamp: now.Add(-1 * time.Hour),
			validate: func(recorded, expected time.Time) bool {
				return recorded.Before(now)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.validate(tt.timestamp, now) {
				t.Errorf("Timestamp validation failed for test %s", tt.name)
			}
		})
	}
}

// TestAuthorizationResultCreation tests AuthorizationResult creation
func TestAuthorizationResultCreation(t *testing.T) {
	tests := []struct {
		name        string
		result      string
		expectError bool
	}{
		{
			name:        "Valid ALLOWED result",
			result:      "ALLOWED",
			expectError: false,
		},
		{
			name:        "Valid DENIED result",
			result:      "DENIED",
			expectError: false,
		},
		{
			name:        "Empty result",
			result:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewAuthorizationResult(tt.result)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for result %s", tt.result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for valid result %s: %v", tt.result, err)
				}
				if result == nil {
					t.Errorf("Expected non-nil result for %s", tt.result)
				} else {
					if result.Value() != tt.result {
						t.Errorf("Expected result value %s, got %s", tt.result, result.Value())
					}
				}
			}
		})
	}
}

// TestDenialReasonCreation tests DenialReason creation
func TestDenialReasonCreation(t *testing.T) {
	tests := []struct {
		name        string
		reason      string
		expectError bool
	}{
		{
			name:        "Valid POLICY_NOT_FOUND reason",
			reason:      "POLICY_NOT_FOUND",
			expectError: false,
		},
		{
			name:        "Valid ROLE_NOT_FOUND reason",
			reason:      "ROLE_NOT_FOUND",
			expectError: false,
		},
		{
			name:        "Valid METHOD_NOT_ALLOWED reason",
			reason:      "METHOD_NOT_ALLOWED",
			expectError: false,
		},
		{
			name:        "Empty reason",
			reason:      "",
			expectError: true,
		},
		{
			name:        "Invalid reason",
			reason:      "INVALID_REASON",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			denialReason, err := NewDenialReason(tt.reason)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for reason %s", tt.reason)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for valid reason %s: %v", tt.reason, err)
				}
				if denialReason == nil {
					t.Errorf("Expected non-nil denial reason for %s", tt.reason)
				} else {
					if denialReason.Value() != tt.reason {
						t.Errorf("Expected reason value %s, got %s", tt.reason, denialReason.Value())
					}
				}
			}
		})
	}
}

// BenchmarkAuditActionCreation benchmarks AuditAction creation
func BenchmarkAuditActionCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewAuditAction("CREATE")
	}
}

// BenchmarkAdminIDCreation benchmarks AdminID creation
func BenchmarkAdminIDCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewAdminID("admin123")
	}
}

// BenchmarkLogStatusCreation benchmarks LogStatus creation
func BenchmarkLogStatusCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewLogStatus("SUCCESS")
	}
}

// BenchmarkAuthorizationResultCreation benchmarks AuthorizationResult creation
func BenchmarkAuthorizationResultCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewAuthorizationResult("ALLOWED")
	}
}
