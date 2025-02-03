package routers

import (
	"regexp/syntax"
)

func isRegexp(input string) bool {
	re, err := syntax.Parse(input, syntax.Perl)
	if err != nil {
		return true
	}

	return !(re.Op == syntax.OpLiteral && len(re.Rune) == len(input))
}
