package gethttps

import (
	"regexp/syntax"
)

func isRegexp(input string) bool {
	re, err := syntax.Parse(input, syntax.Perl)
	if err != nil {
		return true
	}

	return re.Op != syntax.OpLiteral
}

func decodeRegexpEscape(input string) (string, error) {
	re, err := syntax.Parse(input, syntax.PerlX)
	if err != nil {
		return "", err
	}

	return string(re.Rune), nil
}
