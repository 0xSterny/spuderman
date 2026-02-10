package extractor

import (
	"io"
	"path/filepath"
	"strings"
)

type Extractor interface {
	Extract(r io.Reader, filename string) (string, error)
}

// GetExtractor returns the appropriate extractor for the filename
func GetExtractor(filename string) Extractor {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".docx":
		return &DocxExtractor{}
	case ".xlsx":
		return &XlsxExtractor{}
	case ".pdf":
		return &PdfExtractor{}
	case ".doc", ".xls":
		return &StringsExtractor{MinLength: 4}
	case ".txt", ".md", ".ini", ".cfg", ".config", ".ps1", ".sh", ".json", ".xml", ".yaml", ".yml":
		return &TextExtractor{}
	default:
		// Default to TextExtractor for unknown, or Strings?
		// "Other related non-binary files" implies text.
		// Safe bet: Strings Extractor for unknown might be cleaner than dumping binary garbage.
		// But if User says "read specific file types... non-binary", they imply text-like.
		// Let's us TextExtractor for now, assuming user filters mostly text.
		// Actually, let's use Strings for fallback to avoid cluttering matched content with binary noise.
		return &TextExtractor{}
	}
}
