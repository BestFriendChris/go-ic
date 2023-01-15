package ic

import (
	"strings"
	"unicode"
)

func trim(s string) string {
	// only trim when there are newlines
	if strings.IndexRune(s, '\n') == -1 {
		return s
	}

	// Remove first newline
	if s[0] == '\n' {
		s = s[1:]
	}
	minPrefix := -1
	var newLines []string
	for _, line := range strings.Split(s, "\n") {
		i := 0
		for _, c := range line {
			if unicode.IsSpace(c) {
				i += 1
			} else {
				break
			}
		}
		if minPrefix == -1 || i < minPrefix {
			minPrefix = i
		}
		newLines = append(newLines, line)
	}
	var sb strings.Builder
	if minPrefix > 0 {
		for _, line := range newLines {
			line := line[minPrefix:]
			sb.WriteString(line + "\n")
		}
		s = sb.String()
		s = s[:len(s)-1]
	}
	return s
}

func isMultiline(want string) bool {
	return strings.Contains(want, "\n")
}
