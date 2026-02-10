package extractor

import (
	"io"
	"strings"

	"github.com/nguyenthenguyen/docx"
	"github.com/xuri/excelize/v2"
)

// DocxExtractor for .docx files
type DocxExtractor struct{}

func (e *DocxExtractor) Extract(r io.Reader, filename string) (string, error) {
	// docx library needs a ReadAt/Seeker usually?
	// github.com/nguyenthenguyen/docx ReadDocxFromMemory takes *bytes.Reader or io.ReaderAt.
	// We have io.Reader from SMB. We need to buffer it.

	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	rdr := strings.NewReader(string(b)) // Convert to ReaderAt compliant

	d, err := docx.ReadDocxFromMemory(rdr, rdr.Size())
	if err != nil {
		return "", err
	}
	defer d.Close()

	return d.Editable().GetContent(), nil
}

// XlsxExtractor for .xlsx files
type XlsxExtractor struct{}

func (e *XlsxExtractor) Extract(r io.Reader, filename string) (string, error) {
	// excelize.OpenReader works with io.Reader directly?
	// "OpenReader read the spreadsheet from an io.Reader."
	// checking docs: func OpenReader(r io.Reader, opts ...Options) (*File, error)

	f, err := excelize.OpenReader(r)
	if err != nil {
		// Sometimes it might fail if seek is needed, but let's try.
		// If fails, we might need buffering.
		return "", err
	}
	defer f.Close()

	var sb strings.Builder
	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue
		}
		for _, row := range rows {
			for _, col := range row {
				sb.WriteString(col)
				sb.WriteString(" ")
			}
			sb.WriteString("\n")
		}
	}
	return sb.String(), nil
}
