package analytics

import (
	"bytes"
	"testing"
)

func TestRenderPDFStructure(t *testing.T) {
	pdf := renderPDF([]string{"Hello", "World"})
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) {
		t.Fatalf("output missing PDF-1.4 header: got %q", pdf[:20])
	}
	if !bytes.HasSuffix(bytes.TrimRight(pdf, "\n"), []byte("%%EOF")) {
		t.Fatalf("output missing %%EOF trailer")
	}
	// A well-formed xref table opens with "xref\n0 6" for our 5-object doc.
	if !bytes.Contains(pdf, []byte("xref\n0 6")) {
		t.Fatalf("output missing xref table")
	}
}

func TestRenderPDFEscapesParens(t *testing.T) {
	pdf := renderPDF([]string{"Cost is $12 (approx)"})
	// The paren should be backslash-escaped inside the PDF stream.
	if !bytes.Contains(pdf, []byte(`\(approx\)`)) {
		t.Fatalf("parens not escaped: %s", pdf)
	}
}
