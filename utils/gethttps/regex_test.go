package gethttps

import (
	"testing"
)

func TestIsRegexp(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello world\n", false},
		{"\n\t   \r\n", false},
		{"hello*world", true},
		{"\\bword\\b", true},
		{"^start", true},
		{"end$", true},
		{"(group)", true},
		{"abc[def]", true},
		{"", true},
		// NOTE: \Q and \E are used to escape special characters in regex but go does not support it (just ignores them)
		//       once they are supported, this test should be updated
		{`\Qabc\E`, false},
	}

	for _, test := range tests {
		result := isRegexp(test.input)
		if result != test.expected {
			t.Errorf("For input %q, expected %v but got %v", test.input, test.expected, result)
		}
	}
}
