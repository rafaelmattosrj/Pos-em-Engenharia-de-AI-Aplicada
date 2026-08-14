// Package pdfx extrai texto puro de um arquivo PDF — equivalente ao
// ApachePdfBoxDocumentParser usado na versão Java, usando a biblioteca
// pura-Go github.com/ledongthuc/pdf no lugar do Apache PDFBox.
package pdfx

import (
	"bytes"
	"io"

	"github.com/ledongthuc/pdf"
)

// ExtractText lê o PDF em path e retorna todo o texto extraído.
func ExtractText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	reader, err := r.GetPlainText()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return "", err
	}

	return buf.String(), nil
}
