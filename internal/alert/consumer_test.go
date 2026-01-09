package alert

import (
	"testing"
)

func TestShouldTriggerAlert(t *testing.T) {
	tests := []struct {
		name         string
		condition    string
		targetPrice  float64
		currentPrice float64
		expected     bool
	}{
		{
			name:         "ABOVE - triggered",
			condition:    "ABOVE",
			targetPrice:  150.00,
			currentPrice: 151.00,
			expected:     true,
		},
		{
			name:         "ABOVE - exact match",
			condition:    "ABOVE",
			targetPrice:  150.00,
			currentPrice: 150.00,
			expected:     true,
		},
		{
			name:         "ABOVE - not triggered",
			condition:    "ABOVE",
			targetPrice:  150.00,
			currentPrice: 149.99,
			expected:     false,
		},
		{
			name:         "BELOW - triggered",
			condition:    "BELOW",
			targetPrice:  100.00,
			currentPrice: 99.00,
			expected:     true,
		},
		{
			name:         "BELOW - not triggered",
			condition:    "BELOW",
			targetPrice:  100.00,
			currentPrice: 101.00,
			expected:     false,
		},
		{
			name:         "Unknown Condition",
			condition:    "UNKNOWN",
			targetPrice:  100.00,
			currentPrice: 100.00,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldTriggerAlert(tt.condition, tt.targetPrice, tt.currentPrice)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
