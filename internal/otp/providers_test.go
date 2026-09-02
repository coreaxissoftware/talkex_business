package otp

import "testing"

func TestNormalisePhone(t *testing.T) {
	cases := []struct{ in, want string }{
		{"9876543210", "+919876543210"},
		{"+919876543210", "+919876543210"},
		{"919876543210", "+919876543210"},
		{"+1 415 555 0100", "+14155550100"},
		{" +91-9876-54-3210 ", "+919876543210"},
		{"", ""},
	}
	for _, tc := range cases {
		got := normalisePhone(tc.in)
		if got != tc.want {
			t.Errorf("normalisePhone(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
