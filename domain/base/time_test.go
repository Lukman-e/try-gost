package base

import (
	"testing"
)

func TestTimeFields_SetCreateTime(t *testing.T) {
	tests := []struct {
		name string
		att  *TimeFields
	}{
		{
			name: "Test SetTimes",
			att:  &TimeFields{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.att.SetCreateTime()
			if tt.att.CreatedAt == nil || tt.att.UpdatedAt == nil {
				t.Errorf("Expected CreatedAt and UpdatedAt to be set, but one or both are nil")
			}
		})
	}
}

func TestTimeFields_SetUpdateTime(t *testing.T) {
	tests := []struct {
		name string
		att  *TimeFields
	}{
		{
			name: "Test SetUpdateTime",
			att:  &TimeFields{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.att.SetUpdateTime()
			if tt.att.UpdatedAt == nil {
				t.Errorf("Expected UpdatedAt to be set, but it is nil")
			}
		})
	}
}
