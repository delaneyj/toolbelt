package toolbelt

import "strings"

func Pascal(s string) string {
	return toCamel(s, true)
}

func Camel(s string) string {
	return toCamel(s, false)
}

func Snake(s string) string {
	return strings.ToLower(splitAndJoin(s, "_"))
}

func ScreamingSnake(s string) string {
	return strings.ToUpper(splitAndJoin(s, "_"))
}

func Kebab(s string) string {
	return strings.ToLower(splitAndJoin(s, "-"))
}

func Upper(s string) string {
	return strings.ToUpper(s)
}

func Lower(s string) string {
	return strings.ToLower(s)
}

type CasedFn func(string) string

func Cased(s string, fn ...CasedFn) string {
	for _, f := range fn {
		s = f(s)
	}
	return s
}

type CasedString struct {
	Original       string
	Pascal         string
	Camel          string
	Snake          string
	ScreamingSnake string
	Kebab          string
	Upper          string
	Lower          string
}

func ToCasedString(s string) CasedString {
	return CasedString{
		Original:       s,
		Pascal:         Pascal(s),
		Camel:          Camel(s),
		Snake:          Snake(s),
		ScreamingSnake: ScreamingSnake(s),
		Kebab:          Kebab(s),
		Upper:          Upper(s),
		Lower:          Lower(s),
	}
}

// toCamel is adapted to avoid external deps; splits on non-alnum and case boundaries.
func toCamel(s string, upperFirst bool) string {
	parts := splitWords(s)
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i == 0 && !upperFirst {
			parts[i] = strings.ToLower(p[:1]) + strings.ToLower(p[1:])
		} else {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return strings.Join(parts, "")
}

func splitWords(s string) []string {
	var parts []string
	var cur []rune
	for _, r := range s {
		if !(isLetter(r) || isDigit(r)) {
			if len(cur) > 0 {
				parts = append(parts, string(cur))
				cur = cur[:0]
			}
			continue
		}
		if len(cur) > 0 && isBoundary(cur[len(cur)-1], r) {
			parts = append(parts, string(cur))
			cur = cur[:0]
		}
		cur = append(cur, r)
	}
	if len(cur) > 0 {
		parts = append(parts, string(cur))
	}
	return parts
}

func splitAndJoin(s, sep string) string {
	parts := splitWords(s)
	for i, p := range parts {
		parts[i] = strings.ToLower(p)
	}
	return strings.Join(parts, sep)
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
func isLower(r rune) bool { return r >= 'a' && r <= 'z' }

func isBoundary(prev rune, curr rune) bool {
	// Boundary between lower->upper (camelCase) or letter->digit or digit->letter.
	if isLower(prev) && isUpper(curr) {
		return true
	}
	if (isLetter(prev) && isDigit(curr)) || (isDigit(prev) && isLetter(curr)) {
		return true
	}
	return false
}
