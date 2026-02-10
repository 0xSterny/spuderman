package extractor

import (
	"io"
	"strings"
)

// TextExtractor reads everything as string
type TextExtractor struct{}

func (e *TextExtractor) Extract(r io.Reader, filename string) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// StringsExtractor mimics the unix 'strings' command
// It finds printable strings of length >= 4
type StringsExtractor struct {
	MinLength int
}

func (e *StringsExtractor) Extract(r io.Reader, filename string) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	// Scan for sequences of printable chars
	var sb strings.Builder
	runLen := 0
	start := 0

	min := e.MinLength
	if min <= 0 {
		min = 4
	}

	for i, c := range b {
		// isPrint := unicode.IsPrint(rune(c)) && !unicode.IsSpace(rune(c))
		// Better 'strings' logic allows spaces? Yes usually.
		// Let's rely on standard Go IsPrint which includes spaces.
		// But control chars are usually not desired.
		// 'strings' command prints [[:print:]]{4,}

		if c >= 32 && c < 127 { // Simple ASCII printable
			runLen++
		} else {
			if runLen >= min {
				sb.WriteString(string(b[start:i]))
				sb.WriteString("\n")
			}
			runLen = 0
			start = i + 1
		}
	}
	// Flush last run
	if runLen >= min {
		sb.WriteString(string(b[start:]))
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
