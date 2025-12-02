package tmux

import "testing"

func TestSanitizeSessionName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"feature/foo", "feature-foo"},
		{"bugfix_123", "bugfix_123"},
		{"../weird", "weird"},
		{"ABC:def", "ABCdef"},
		{"", ""},
	}

	for _, c := range cases {
		if got := sanitizeSessionName(c.in); got != c.want {
			t.Fatalf("sanitizeSessionName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
