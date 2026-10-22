package convert

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLongRailUsesFlowingDataTable(t *testing.T) {
	p := func(text string, h float64) Block {
		return Block{Kind: "paragraph", Style: Style{Height: h}, Runs: []Run{{Text: text}}}
	}
	table := Block{Kind: "table", Tag: "table", ID: "data", Style: Style{Width: 493}}
	for i := 0; i < 35; i++ {
		table.Rows = append(table.Rows, Row{Header: i == 0, Cells: []Cell{{Span: 1, Blocks: []Block{p("Cell", 30)}}}})
	}
	doc := Document{Blocks: []Block{{Kind: "table", Tag: "layout", Grid: []float64{130, 20, 493}, Rows: []Row{{Cells: []Cell{{Span: 1, Blocks: []Block{p("Rail", 100)}}, {Span: 1}, {Span: 1, Blocks: []Block{p("Introduction", 300), table, p("Conclusion", 50)}}}}}}}}
	data, err := Build(context.Background(), &doc, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, b := range doc.Blocks {
		if b.ID == "data" {
			found = true
			if b.Kind != "table" || b.Style.Margin[3] != 150 {
				t.Fatal("data table no longer aligned with its reading column")
			}
		}
	}
	if !found {
		t.Fatal("long data table remains trapped inside a tall layout row")
	}
	if strings.Count(unzip(t, data)["word/document.xml"], ">Cell</w:t>") != 35 {
		t.Fatal("data rows lost during rail adaptation")
	}
}

func TestDocumentFooterMetadata(t *testing.T) {
	if testing.Short() || os.Getenv("HTMLWORD_INTEGRATION") != "1" {
		t.Skip("browser integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	doc, err := Analyze(ctx, `<p>Body content</p><footer style="font-size:9px;display:flex;justify-content:space-between"><span>Document metadata</span><a href="https://example.com">Reference</a></footer>`, "", DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	data, err := Build(ctx, &doc, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	f := unzip(t, data)
	if strings.Contains(f["word/document.xml"], "Document metadata") || !strings.Contains(f["word/footer1.xml"], "Document metadata") {
		t.Fatal("short document footer must not become orphaned body content")
	}
	if !strings.Contains(f["word/_rels/footer1.xml.rels"], "https://example.com") {
		t.Fatal("footer hyperlinks require relationships scoped to the footer part")
	}
}

func TestCompactMixedHeadingRowStaysTogether(t *testing.T) {
	if testing.Short() || os.Getenv("HTMLWORD_INTEGRATION") != "1" {
		t.Skip("browser integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	doc, err := Analyze(ctx, `<div style="display:grid;grid-template-columns:2fr 1fr;gap:20px"><div><h1>Project quotation</h1><p>Scope and delivery</p></div><div style="background:#eee;padding:20px"><p>Reference 032</p><p>Revision 3</p></div></div>`, "", DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range doc.Blocks {
		if b.Tag == "layout" && len(b.Rows) == 1 && b.Rows[0].Keep {
			return
		}
	}
	t.Fatal("a compact heading and its painted metadata panel may split when paragraphs are inserted before them")
}
