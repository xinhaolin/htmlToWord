package convert

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xinhaolin/htmlToWord/internal/cdp"
)

func unzip(t *testing.T, data []byte) map[string]string {
	t.Helper()
	z, e := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if e != nil {
		t.Fatal(e)
	}
	files := map[string]string{}
	for _, f := range z.File {
		r, e := f.Open()
		if e != nil {
			t.Fatal(e)
		}
		b, e := io.ReadAll(r)
		r.Close()
		if e != nil {
			t.Fatal(e)
		}
		files[f.Name] = string(b)
		if strings.HasSuffix(f.Name, ".xml") || strings.HasSuffix(f.Name, ".rels") {
			decoder := xml.NewDecoder(bytes.NewReader(b))
			for {
				_, e = decoder.Token()
				if e == io.EOF {
					break
				}
				if e != nil {
					t.Fatalf("invalid XML %s: %v", f.Name, e)
				}
			}
		}
	}
	return files
}
func TestNativeDocument(t *testing.T) {
	opts := DefaultOptions()
	opts.Header = "测试页眉"
	opts.Footer = "测试页脚"
	opts.PageNumbers = true
	style := Style{Size: 16, Font: "Arial", Color: "234567", LineHeight: 24}
	doc := Document{Title: "Editable <document>", Blocks: []Block{
		{Kind: "paragraph", Level: 1, ID: "intro", Style: style, Runs: []Run{{Text: "标题 & 安全 <文本>", Style: style}}},
		{Kind: "paragraph", Style: style, ListID: 1, Ordered: true, Start: 3, Runs: []Run{{Text: "项目一", Style: style}, {Text: "网站", Style: style, Link: "https://example.com?a=1&b=2"}}},
		{Kind: "paragraph", Style: style, Runs: []Run{{Text: "回到标题", Link: "#intro", Style: style}, {Text: "行一\n\t行二", Style: style}}},
		{Kind: "table", Rows: []Row{{Header: true, Cells: []Cell{{Span: 2, Blocks: []Block{{Kind: "paragraph", Runs: []Run{{Text: "合并列"}}}}}}}, {Cells: []Cell{{Span: 1, RowSpan: 2, Blocks: []Block{{Kind: "paragraph", Runs: []Run{{Text: "纵向合并"}}}}}, {Span: 1, Blocks: []Block{{Kind: "paragraph", Runs: []Run{{Text: "第一行"}}}}}}}, {Cells: []Cell{{Span: 1, Blocks: []Block{{Kind: "paragraph", Runs: []Run{{Text: "第二行"}}}}}}}}},
	}}
	data, e := Build(context.Background(), &doc, opts)
	if e != nil {
		t.Fatal(e)
	}
	f := unzip(t, data)
	body := f["word/document.xml"]
	for _, s := range []string{"<w:p>", "<w:numPr>", "<w:tbl>", "<w:tblHeader/>", `w:val="restart"`, `w:val="continue"`, `w:val="2"`, `<w:hyperlink r:id=`, `<w:hyperlink w:anchor=`, `<w:bookmarkStart`, `<w:br/>`, `<w:tab/>`, `&amp; 安全 &lt;文本&gt;`} {
		if !strings.Contains(body, s) {
			t.Errorf("missing %s", s)
		}
	}
	for _, s := range []string{"altChunk", "txbxContent", "wp:anchor", "documentProtection"} {
		if strings.Contains(body, s) {
			t.Errorf("unexpected fixed/protected content %s", s)
		}
	}
	if !strings.Contains(f["word/numbering.xml"], `<w:start w:val="3"/>`) {
		t.Error("list start lost")
	}
	if !strings.Contains(f["word/footer1.xml"], `w:instr="NUMPAGES"`) {
		t.Error("page field lost")
	}
}
func TestImageEmbedding(t *testing.T) {
	// A complete 1x1 PNG, reused twice to verify relationship/media deduplication.
	source := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII="
	doc := Document{Blocks: []Block{{Kind: "paragraph", Runs: []Run{{Image: &Picture{Source: source, Width: 80, Height: 80, Alt: "test"}}, {Image: &Picture{Source: source, Width: 80, Height: 80, Alt: "test"}}}}}}
	data, e := Build(context.Background(), &doc, DefaultOptions())
	if e != nil {
		t.Fatal(e)
	}
	f := unzip(t, data)
	if len(doc.Warnings) > 0 {
		t.Fatal(doc.Warnings)
	}
	if _, ok := f["word/media/image1.png"]; !ok {
		t.Error("embedded image missing")
	}
	if _, ok := f["word/media/image2.png"]; ok {
		t.Error("media was not deduplicated")
	}
	if strings.Count(f["word/document.xml"], "<w:drawing>") != 2 {
		t.Error("drawing missing")
	}
}
func TestValidation(t *testing.T) {
	for _, raw := range []string{"http://127.0.0.1/x", "http://[::1]/x", "http://192.168.1.1/x", "file:///etc/passwd", "https://example.com:8080/a"} {
		if PublicResource(context.Background(), raw) == nil {
			t.Errorf("allowed %s", raw)
		}
	}
	for _, name := range []string{"../../bad\r\n.docx", "<script>.docx", ""} {
		out := SafeFilename(name)
		if strings.ContainsAny(out, "/\\\r\n<>") || !strings.HasSuffix(out, ".docx") {
			t.Errorf("unsafe %q", out)
		}
	}
	opts := DefaultOptions()
	opts.MarginMM = -1
	if opts.Validate() == nil {
		t.Error("accepted invalid margin")
	}
}

func TestColumnSections(t *testing.T) {
	opts := DefaultOptions()
	opts.Header = "Header"
	opts.PageNumbers = true
	doc := Document{Blocks: []Block{{Kind: "columns", Columns: 2, Gap: 20, Children: []Block{{Kind: "paragraph", Runs: []Run{{Text: "Column text"}}}}}, {Kind: "table", Rows: []Row{{Cells: []Cell{{Span: 1, Blocks: []Block{{Kind: "columns", Columns: 2, Children: []Block{{Kind: "paragraph", Runs: []Run{{Text: "Nested content"}}}}}}}}}}}}}
	data, e := Build(context.Background(), &doc, opts)
	if e != nil {
		t.Fatal(e)
	}
	f := unzip(t, data)
	body := f["word/document.xml"]
	if strings.Count(body, "<w:sectPr>") != 3 || strings.Count(body, "<w:headerReference") != 3 || strings.Count(body, "<w:footerReference") != 3 {
		t.Error("sections must consistently reference headers and footers")
	}
	if !strings.Contains(body, "Nested content") || len(doc.Warnings) == 0 {
		t.Error("nested columns must preserve text and report adaptation")
	}
	if !strings.Contains(body[strings.LastIndex(body, "<w:sectPr>"):], `<w:type w:val="continuous"/>`) {
		t.Error("returning from CSS columns must not introduce a new-page section")
	}
}
func TestBrowserConversion(t *testing.T) {
	if testing.Short() || os.Getenv("HTMLWORD_INTEGRATION") != "1" {
		t.Skip("browser integration")
	}
	if _, e := cdp.FindChrome(); e != nil {
		t.Skip(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	request := Request{Options: DefaultOptions(), HTML: `<!doctype html><html><head><title>Integration</title><style>body{font-family:Arial;font-size:16px}.accent{color:rgb(12,34,56);font-weight:bold}.hide{display:none}.row{display:flex;gap:12px}.row>div{flex:1;padding:8px;background:#eee}td{border:1px solid black;padding:6px}</style></head><body><script>document.body.innerHTML='EXECUTED';</script><h1>标题</h1><p>hello <span class="accent">styled</span> world</p><p class="hide">HIDDEN</p><ol start="4"><li>one<ul><li>nested</li></ul></li><li>two</li></ol><table><tr><td rowspan="2">A</td><td>B</td></tr><tr><td>C</td></tr></table><div class="row"><div><p>Left</p></div><div><p>Right</p></div></div><p style="break-before:page">New page</p></body></html>`}
	result, e := Convert(ctx, request)
	if e != nil {
		t.Fatal(e)
	}
	files := unzip(t, result.Data)
	body := files["word/document.xml"]
	for _, s := range []string{"标题", "styled", "hello ", " world", "nested", "Left", "Right", "New page", `w:val="0C2238"`, `<w:pageBreakBefore/>`, `w:val="continue"`} {
		if !strings.Contains(body, s) {
			t.Errorf("lost content/style %q", s)
		}
	}
	if strings.Contains(body, "HIDDEN") || strings.Contains(body, "EXECUTED") {
		t.Error("hidden/active content exported")
	}
	if result.Document.Stats.Tables != 2 || result.Document.Stats.Lists != 2 {
		t.Errorf("unexpected stats: %+v", result.Document.Stats)
	}
	if !strings.Contains(files["word/numbering.xml"], `w:val="4"`) {
		t.Error("HTML ordered-list start lost")
	}
}

// A centered portrait and mixed inline/block flex items must remain in the same
// editable row, including when that row is nested inside another layout cell.
func TestNestedAlignedLayouts(t *testing.T) {
	if testing.Short() || os.Getenv("HTMLWORD_INTEGRATION") != "1" {
		t.Skip("browser integration")
	}
	if _, e := cdp.FindChrome(); e != nil {
		t.Skip(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	html := `<!doctype html><style>*{box-sizing:border-box}body{margin:0;font:12px/1.7 Arial}.outer{display:grid;grid-template-columns:120px 1fr}.header{display:grid;grid-template-columns:1fr 60px;gap:20px;align-items:center;padding:20px;background:#123456}.header h1{margin:0;font-size:28px}.header p{margin:0}.label{display:flex;align-items:baseline;gap:9px;border-bottom:1px solid #123456;margin-bottom:13px}.label h2{margin:0;font-size:14px}.date{display:flex;justify-content:space-between;align-items:baseline;gap:9px}.date h3{margin:0}.track{height:4px;background:#ddd;margin-top:6px}.fill{display:block;height:4px;width:76%;background:#448877}.skill-label{display:flex;justify-content:space-between;font-size:9px}.skill-label span:last-child{font-size:8px}.contact-title{font-size:11px;margin:0 0 13px}.contact-title small{display:block;font-size:7px;margin-top:2px}</style><div class="header" id="header"><div><h1>Name</h1><p>Role</p><p>Intent</p><p>Additional line</p></div><img width="60" height="50" src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII="></div><div class="outer"><aside><h2 class="contact-title" id="contact-title">Contact<small>CONTACT</small></h2><div class="skill-label" id="skill-label"><span>TypeScript / JavaScript</span><span>94%</span></div><div class="track" id="track"><span class="fill"></span></div></aside><main><div class="label" id="label"><b>01</b><h2>Profile</h2><span>ENGLISH</span></div><div class="date" id="date"><h3>Company</h3><span>2022 — 2026</span></div></main></div>`
	doc, e := Analyze(ctx, html, "", DefaultOptions())
	if e != nil {
		t.Fatal(e)
	}
	var found = map[string]Block{}
	var visit func([]Block)
	visit = func(bs []Block) {
		for _, b := range bs {
			if b.ID != "" {
				found[b.ID] = b
			}
			visit(b.Children)
			for _, r := range b.Rows {
				for _, c := range r.Cells {
					visit(c.Blocks)
				}
			}
		}
	}
	visit(doc.Blocks)
	var text func([]Block) string
	text = func(bs []Block) string {
		var out strings.Builder
		for _, b := range bs {
			for _, r := range b.Runs {
				out.WriteString(r.Text)
			}
			out.WriteString(text(b.Children))
			for _, r := range b.Rows {
				for _, c := range r.Cells {
					out.WriteString(text(c.Blocks))
				}
			}
		}
		return out.String()
	}
	var layoutRows func(Block) []Row
	layoutRows = func(b Block) []Row {
		if b.Tag == "layout" {
			return b.Rows
		}
		for _, r := range b.Rows {
			for _, c := range r.Cells {
				for _, child := range c.Blocks {
					if rows := layoutRows(child); len(rows) > 0 {
						return rows
					}
				}
			}
		}
		return nil
	}
	for _, id := range []string{"header", "label", "date", "skill-label"} {
		b, ok := found[id]
		if !ok {
			t.Errorf("lost layout id %s", id)
			continue
		}
		if id == "label" || id == "date" || id == "skill-label" {
			if b.Kind != "paragraph" {
				t.Errorf("%s must keep one editable text paragraph, got %s", id, b.Kind)
			}
			if id == "label" && strings.ReplaceAll(text([]Block{b}), " ", "") != "01ProfileENGLISH" {
				t.Errorf("inline/block flex order changed: %q", text([]Block{b}))
			}
			if (id == "date" || id == "skill-label") && (len(b.Tabs) != 1 || b.Tabs[0].Align != "right") {
				t.Error("date must use a native right-aligned tab")
			}
			continue
		}
		rows := layoutRows(b)
		if len(rows) != 1 {
			t.Errorf("%s must have one row, got %d", id, len(rows))
			continue
		}
		if id == "label" && text([]Block{b}) != "01ProfileENGLISH" {
			t.Errorf("inline/block flex order changed: %q", text([]Block{b}))
		}
		if id == "header" {
			images := 0
			for _, c := range rows[0].Cells {
				for _, p := range c.Blocks {
					for _, r := range p.Runs {
						if r.Image != nil {
							images++
						}
					}
				}
			}
			if images != 1 {
				t.Errorf("portrait must be in the centered header row, got %d", images)
			}
		}
	}
	if title := found["contact-title"]; title.Style.Margin[2] != 0 {
		t.Error("parent heading bottom margin must not be applied before the subtitle")
	}
	bar, ok := found["track"]
	if !ok || bar.Kind != "table" || len(bar.Rows) != 1 || bar.Rows[0].Height != 4 || !bar.Rows[0].Exact {
		t.Errorf("empty skill track must have exact 4px height, got %+v", bar)
	}
	if len(bar.Rows) > 0 && len(bar.Rows[0].Cells) >= 2 {
		fill, rest := bar.Rows[0].Cells[0].Style.Width, bar.Rows[0].Cells[1].Style.Width
		if fill/(fill+rest) < .75 || fill/(fill+rest) > .77 {
			t.Errorf("76%% fill lost: %v / %v", fill, rest)
		}
	} else {
		t.Error("track must preserve fill and remaining background")
	}
	data, e := Build(ctx, &doc, DefaultOptions())
	if e != nil {
		t.Fatal(e)
	}
	body := unzip(t, data)["word/document.xml"]
	if !strings.Contains(body, `<w:trHeight w:val="60" w:hRule="exact"/>`) {
		t.Error("4px decorative row must remain thin in DOCX")
	}
	if !strings.Contains(body, `<w:tab w:val="right"`) {
		t.Error("right aligned date tab missing from DOCX")
	}
}

func TestSmallParagraphMark(t *testing.T) {
	s := Style{Size: 7, Font: "Arial", LineHeight: 12}
	doc := Document{Blocks: []Block{{Kind: "paragraph", Style: s, Runs: []Run{{Text: "small sidebar label", Style: s}}}}}
	data, e := Build(context.Background(), &doc, DefaultOptions())
	if e != nil {
		t.Fatal(e)
	}
	body := unzip(t, data)["word/document.xml"]
	before, _, _ := strings.Cut(body, "</w:pPr>")
	if !strings.Contains(before, `<w:sz w:val="11"/>`) {
		t.Error("paragraph mark must use the small paragraph's font size instead of Normal 11pt")
	}
}

// A CSS heading must stay connected through its layout wrappers and the
// following table's margin, without locking ordinary body content together.
func TestCSSPaginationThroughLayoutTables(t *testing.T) {
	if testing.Short() || os.Getenv("HTMLWORD_INTEGRATION") != "1" {
		t.Skip("browser integration")
	}
	if _, e := cdp.FindChrome(); e != nil {
		t.Skip(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	html := `<!doctype html><style>body{margin:0;font:12px Arial}p,h2{margin:0}.heading{padding:6px;background:#eee;break-after:avoid-page}.heading>div{display:flex;gap:8px}.card{padding:6px;break-inside:avoid}</style><div class="heading"><div><b>01</b><h2>HEADING</h2></div></div><table style="margin-top:17px"><tr><td>BODY</td></tr></table><section class="card" id="card"><p>CARD</p><p>DETAIL</p></section><p>AFTER</p>`
	doc, e := Analyze(ctx, html, "", DefaultOptions())
	if e != nil {
		t.Fatal(e)
	}
	cardFound := false
	for _, b := range doc.Blocks {
		if b.ID == "card" {
			cardFound = b.Kind == "table" && b.Style.Keep
		}
	}
	if !cardFound {
		t.Error("break-inside:avoid must survive a painted container's layout mapping")
	}
	data, e := Build(ctx, &doc, DefaultOptions())
	if e != nil {
		t.Fatal(e)
	}
	body := unzip(t, data)["word/document.xml"]
	decoder := xml.NewDecoder(strings.NewReader(body))
	var inP, inText, keep, sawBody bool
	var text strings.Builder
	chained := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		switch v := token.(type) {
		case xml.StartElement:
			switch v.Name.Local {
			case "p":
				inP, keep = true, false
				text.Reset()
			case "keepNext":
				keep = inP
			case "t":
				inText = true
			}
		case xml.CharData:
			if inText {
				text.Write(v)
			}
		case xml.EndElement:
			if v.Name.Local == "t" {
				inText = false
			}
			if v.Name.Local == "p" {
				if strings.Contains(text.String(), "BODY") {
					sawBody = true
				}
				if !sawBody {
					chained++
					if !keep {
						t.Errorf("heading chain broken by paragraph %q", text.String())
					}
				} else if text.String() == "CARD" || text.String() == "DETAIL" {
					if !keep {
						t.Error("card contents must remain connected through the final spacing paragraph")
					}
				} else if keep {
					t.Errorf("ordinary content must remain free to reflow: %q", text.String())
				}
				inP = false
			}
		}
	}
	if !sawBody || chained < 3 || strings.Contains(body, "<w:cantSplit/>") {
		t.Error("missing native pagination constraints or fixture content")
	}
}
