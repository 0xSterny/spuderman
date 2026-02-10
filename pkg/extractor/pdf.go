package extractor

import (
	"bytes"
	"io"

	"github.com/ledongthuc/pdf"
)

type PdfExtractor struct{}

func (e *PdfExtractor) Extract(r io.Reader, filename string) (string, error) {
	// ledongthuc/pdf needs io.ReaderAt and size.
	// We have to buffer it.
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	rdr := bytes.NewReader(b)

	f, err := pdf.NewReader(rdr, rdr.Size())
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	// Get total pages
	totalPage := f.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := f.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		s, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(s)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}
