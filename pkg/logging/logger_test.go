package logging

import "testing"

func TestIsSensitiveKey(t *testing.T) {
	cases := []struct {
		in  string
		out bool
	}{
		{"password", true},
		{"passwd", true},
		{"secret", true},
		{"token", true},
		{"authorization", true},
		{"api_key", true},
		{"apikey", true},
		{"cookie", true},
		{"user_id", false},
		{"password_hash", false}, // suffix not matching "_password" exactly
		{"my_password", true},    // matches suffix rule
	}

	for _, c := range cases {
		got := isSensitiveKey(c.in)
		if got != c.out {
			t.Fatalf("isSensitiveKey(%q) = %v; want %v", c.in, got, c.out)
		}
	}
}
