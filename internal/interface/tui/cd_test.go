package tui

import "testing"

func TestParseDigit(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"0", 0, false},
		{"5", 5, false},
		{"a", 0, true},
		{"12", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		got, err := parseDigit(tt.in)
		if tt.wantErr && err == nil {
			t.Fatalf("parseDigit(%q) expected error", tt.in)
		}
		if !tt.wantErr && (err != nil || got != tt.want) {
			t.Fatalf("parseDigit(%q) = (%d, %v), want (%d, nil)", tt.in, got, err, tt.want)
		}
	}
}
