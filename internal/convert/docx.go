package convert

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const wordNS = `xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture" xmlns:w14="http://schemas.microsoft.com/office/word/2010/wordml" xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006" mc:Ignorable="w14"`
const xmlHead = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`
const relBase = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/"

type relation struct {
	ID, Kind, Target string
	External         bool
}
type media struct {
	Name          string
	Data          []byte
	Width, Height int
}
type listSpec struct {
	ID      int
	Ordered bool
	Start   int
	Format  string
}
type writer struct {
	ctx             context.Context
	options         Options
	doc             *Document
	rels            []relation
	media           []media
	lists           map[int]listSpec
	imageCache      map[string]int
	bookmarks       map[string]int
	drawingID       int
	bookmarkID      int
	headerID        string
	footerID        string
	charts          int
	controlID       int
	extraParts      map[string][]byte
	err             error
	continuousFinal bool
	execution       Runtime
}

func esc(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(strings.Map(func(r rune) rune {
		if r == 9 || r == 10 || r == 13 || r >= 32 && r != 0xfffe && r != 0xffff {
			return r
		}
		return -1
	}, s)))
	return b.String()
}
func integer(x float64) string      { return strconv.Itoa(int(math.Round(x))) }
func twip(px float64) string        { return integer(px * 15) }
func val(tag, value string) string  { return `<w:` + tag + ` w:val="` + esc(value) + `"/>` }
func nonnegative(x float64) float64 { return math.Max(0, x) }
func (w *writer) warning(s string) {
	code := "LAYOUT_APPROXIMATED"
	switch {
	case strings.HasPrefix(s, "表格或嵌套布局中的多栏"):
		code = "NESTED_COLUMNS_FLATTENED"
	case strings.HasPrefix(s, "图片超过"):
		code = "IMAGE_LIMIT"
	case strings.HasPrefix(s, "图片转换失败"):
		code = "IMAGE_UNAVAILABLE"
	case strings.HasPrefix(s, "长侧栏布局"):
		code = "SIDEBAR_ADAPTED"
	}
	w.doc.Diagnostics = append(w.doc.Diagnostics, Diagnostic{Code: code, Message: s, Stage: "docx"})
	for _, v := range w.doc.Warnings {
		if v == s {
			return
		}
	}
	w.doc.Warnings = append(w.doc.Warnings, s)
}
func (w *writer) addRel(kind, target string, external bool) string {
	for _, r := range w.rels {
		if r.Kind == kind && r.Target == target && r.External == external {
			return r.ID
		}
	}
	id := fmt.Sprintf("rId%d", len(w.rels)+1)
	w.rels = append(w.rels, relation{id, kind, target, external})
	return id
}

func Convert(ctx context.Context, request Request) (Result, error) {
	start := time.Now()
	if e := request.Options.Validate(); e != nil {
		return Result{}, e
	}
	if strings.TrimSpace(request.HTML) == "" {
		return Result{}, fmt.Errorf("请输入 HTML")
	}
	if len(request.HTML)+len(request.CSS) > 8<<20 {
		return Result{}, fmt.Errorf("HTML 和 CSS 合计不能超过 8 MiB")
	}
	doc, e := Analyze(ctx, request.HTML, request.CSS, request.Options)
	if e != nil {
		return Result{}, e
	}
	data, e := Build(ctx, &doc, request.Options)
	if e != nil {
		return Result{}, e
	}
	return Result{Data: data, Document: doc, DurationMS: time.Since(start).Milliseconds()}, nil
}

// Build writes native WordprocessingML. No HTML import parts, fixed line boxes,
// PDF conversion, document protection, or screenshot text overlays are used.
func Build(ctx context.Context, doc *Document, o Options) ([]byte, error) {
	return BuildWithRuntime(ctx, doc, o, Runtime{})
}

func BuildWithRuntime(ctx context.Context, doc *Document, o Options, execution Runtime) ([]byte, error) {
	if e := o.Validate(); e != nil {
		return nil, e
	}
	w := &writer{ctx: ctx, options: o, doc: doc, lists: map[int]listSpec{}, imageCache: map[string]int{}, bookmarks: map[string]int{}, extraParts: map[string][]byte{}}
	w.execution = execution
	doc.Blocks = w.paginate(doc.Blocks)
	w.addRel("styles", "styles.xml", false)
	w.addRel("settings", "settings.xml", false)
	w.addRel("fontTable", "fontTable.xml", false)
	w.addRel("numbering", "numbering.xml", false)
	var collect func([]Block)
	collect = func(bs []Block) {
		for _, b := range bs {
			if b.ID != "" {
				if _, ok := w.bookmarks[b.ID]; !ok {
					w.bookmarkID++
					w.bookmarks[b.ID] = w.bookmarkID
				}
			}
			if b.ListID > 0 {
				if _, ok := w.lists[b.ListID]; !ok {
					w.lists[b.ListID] = listSpec{b.ListID, b.Ordered, max(1, b.Start), b.ListFormat}
				}
			}
			collect(b.Children)
			for _, r := range b.Rows {
				for _, c := range r.Cells {
					collect(c.Blocks)
				}
			}
		}
	}
	collect(doc.Blocks)
	headerID, footerID := "", ""
	if o.Header != "" {
		headerID = w.addRel("header", "header1.xml", false)
	}
	if o.Footer != "" || o.PageNumbers || len(doc.Footer) > 0 {
		footerID = w.addRel("footer", "footer1.xml", false)
	}
	w.headerID, w.footerID = headerID, footerID
	content := w.blocks(doc.Blocks, o.ContentPX(), 0)
	if w.err != nil {
		return nil, w.err
	}
	document := xmlHead + `<w:document ` + wordNS + `><w:body>` + content + w.section(1, 0, headerID, footerID, w.continuousFinal) + `</w:body></w:document>`
	createdAt := time.Now().UTC()
	if execution.CreatedAt != nil {
		createdAt = execution.CreatedAt.UTC()
	}
	entries := map[string][]byte{
		"[Content_Types].xml": []byte(w.contentTypes(headerID != "", footerID != "")),
		"_rels/.rels":         []byte(xmlHead + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="` + relBase + `officeDocument" Target="word/document.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/><Relationship Id="rId3" Type="` + relBase + `extended-properties" Target="docProps/app.xml"/></Relationships>`),
		"word/document.xml":   []byte(document), "word/styles.xml": []byte(w.styles()), "word/settings.xml": []byte(xmlHead + `<w:settings ` + wordNS + `><w:defaultTabStop w:val="720"/><w:autoHyphenation w:val="false"/><w:compat><w:compatSetting w:name="compatibilityMode" w:uri="http://schemas.microsoft.com/office/word" w:val="15"/></w:compat></w:settings>`),
		"word/fontTable.xml": []byte(xmlHead + `<w:fonts ` + wordNS + `><w:font w:name="` + esc(o.Font) + `"><w:family w:val="swiss"/></w:font></w:fonts>`),
		"word/numbering.xml": []byte(w.numbering()), "word/_rels/document.xml.rels": []byte(w.relationships()),
		"docProps/core.xml": []byte(xmlHead + `<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><dc:title>` + esc(doc.Title) + `</dc:title><dc:creator>HTMLWord</dc:creator><dcterms:created xsi:type="dcterms:W3CDTF">` + createdAt.Format(time.RFC3339) + `</dcterms:created></cp:coreProperties>`),
		"docProps/app.xml":  []byte(xmlHead + `<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"><Application>HTMLWord Go</Application></Properties>`),
	}
	if headerID != "" {
		entries["word/header1.xml"] = []byte(w.headerFooter("hdr", o.Header, false))
	}
	if footerID != "" {
		entries["word/footer1.xml"] = []byte(w.headerFooter("ftr", o.Footer, o.PageNumbers))
	}
	for _, m := range w.media {
		entries["word/media/"+m.Name] = m.Data
	}
	for name, data := range w.extraParts {
		entries[name] = data
	}
	var output bytes.Buffer
	z := zip.NewWriter(boundedOutput{buffer: &output, limit: execution.MaxOutputBytes})
	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if e := ctx.Err(); e != nil {
			z.Close()
			return nil, e
		}
		f, e := z.Create(name)
		if e != nil {
			return nil, e
		}
		if _, e = f.Write(entries[name]); e != nil {
			return nil, e
		}
	}
	if e := z.Close(); e != nil {
		return nil, e
	}
	return output.Bytes(), nil
}

var ErrOutputLimit = errors.New("DOCX output limit exceeded")

type boundedOutput struct {
	buffer *bytes.Buffer
	limit  int64
}

func (b boundedOutput) Write(p []byte) (int, error) {
	if b.limit > 0 && int64(b.buffer.Len())+int64(len(p)) > b.limit {
		return 0, ErrOutputLimit
	}
	return b.buffer.Write(p)
}

func (w *writer) blocks(blocks []Block, width float64, depth int) string {
	return w.blocksKeeping(blocks, width, depth, false)
}

// A painted heading can be represented by several nested layout tables. Keep
// its content and mandatory spacing paragraphs connected to the following block.
func (w *writer) blocksKeeping(blocks []Block, width float64, depth int, keepNext bool) string {
	if err := w.ctx.Err(); err != nil {
		w.err = err
		return ""
	}
	if depth > 80 {
		w.err = errors.New("HTMLWORD_LIMIT: native document nesting exceeds 80")
		return ""
	}
	var out strings.Builder
	previousKeepNext := keepNext
	for _, b := range blocks {
		b.Style.KeepNext = b.Style.KeepNext || keepNext
		switch b.Kind {
		case "table":
			if b.Style.BreakBefore {
				out.WriteString(`<w:p><w:pPr><w:pageBreakBefore/></w:pPr></w:p>`)
			}
			if b.Style.Margin[0] > 0 {
				out.WriteString(spacingParagraph(b.Style.Margin[0], previousKeepNext || b.Style.KeepNext))
			}
			out.WriteString(w.table(b, width, depth+1))
			out.WriteString(spacingParagraph(b.Style.Margin[2], b.Style.KeepNext))
			if b.Style.BreakAfter {
				out.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)
			}
		case "columns":
			if depth > 0 {
				w.warning("表格或嵌套布局中的多栏已展开为连续段落，以保留自然编辑能力。")
				out.WriteString(w.blocksKeeping(b.Children, width, depth+1, b.Style.KeepNext))
				continue
			}
			out.WriteString(`<w:p><w:pPr>` + w.section(1, 0, "", "", true) + `</w:pPr></w:p>`)
			out.WriteString(w.blocksKeeping(b.Children, math.Max(24, (width-b.Gap*float64(max(0, b.Columns-1)))/float64(max(1, b.Columns))), depth+1, b.Style.KeepNext))
			out.WriteString(`<w:p><w:pPr>` + w.section(max(1, b.Columns), b.Gap, "", "", true) + `</w:pPr></w:p>`)
			w.continuousFinal = true
		case "rule":
			out.WriteString(`<w:p><w:pPr><w:pBdr><w:bottom w:val="single" w:sz="6" w:color="` + fallback(b.Style.Borders[0].Color, "D1D5DB") + `"/></w:pBdr><w:spacing w:before="` + twip(b.Style.Margin[0]) + `" w:after="` + twip(b.Style.Margin[2]) + `"/></w:pPr></w:p>`)
		default:
			out.WriteString(w.paragraph(b, width))
		}
		previousKeepNext = b.Style.KeepNext || b.Level > 0
	}
	return out.String()
}

// Word requires a paragraph after a nested table. Give its paragraph mark a
// minimal line box instead of inheriting the document's 11pt default font.
func spacingParagraph(after float64, keepNext bool) string {
	keep := ""
	if keepNext {
		keep = `<w:keepNext/>`
	}
	return `<w:p><w:pPr>` + keep + `<w:spacing w:before="0" w:after="` + twip(nonnegative(after)) + `" w:line="1" w:lineRule="exact"/><w:rPr><w:sz w:val="2"/><w:szCs w:val="2"/></w:rPr></w:pPr></w:p>`
}
func fallback(s, defaultValue string) string {
	if s == "" {
		return defaultValue
	}
	return s
}
func (w *writer) paragraph(b Block, width float64) string {
	s := b.Style
	var p strings.Builder
	p.WriteString(`<w:p><w:pPr>`)
	if b.Level > 0 {
		p.WriteString(val("pStyle", fmt.Sprintf("Heading%d", b.Level)))
	}
	if b.Level > 0 || s.KeepNext {
		p.WriteString(`<w:keepNext/>`)
	}
	if s.Keep {
		p.WriteString(`<w:keepLines/>`)
	}
	if s.NoWrap {
		p.WriteString(`<w:wordWrap w:val="false"/>`)
	}
	if len(b.Tabs) > 0 {
		p.WriteString(`<w:tabs>`)
		for _, tab := range b.Tabs {
			p.WriteString(`<w:tab w:val="` + esc(tab.Align) + `" w:pos="` + twip(tab.Position) + `"/>`)
		}
		p.WriteString(`</w:tabs>`)
	}
	if s.BreakBefore {
		p.WriteString(`<w:pageBreakBefore/>`)
	}
	p.WriteString(`<w:widowControl/>`)
	if b.ListID > 0 {
		p.WriteString(`<w:numPr><w:ilvl w:val="` + strconv.Itoa(min(8, max(0, b.ListLevel))) + `"/><w:numId w:val="` + strconv.Itoa(b.ListID) + `"/></w:numPr>`)
		if s.Margin[3]+s.Padding[3] > 0 {
			p.WriteString(`<w:tabs><w:tab w:val="clear" w:pos="` + strconv.Itoa(360*(b.ListLevel+1)) + `"/><w:tab w:val="num" w:pos="` + twip(s.Margin[3]+s.Padding[3]) + `"/></w:tabs>`)
		}
	}
	p.WriteString(paragraphBorders(s.Borders))
	if s.Background != "" {
		p.WriteString(`<w:shd w:val="clear" w:color="auto" w:fill="` + esc(s.Background) + `"/>`)
	}
	if s.Direction == "rtl" {
		p.WriteString(`<w:bidi/>`)
	}
	before, after := nonnegative(s.Margin[0]+s.Padding[0]), nonnegative(s.Margin[2]+s.Padding[2])
	p.WriteString(`<w:spacing w:before="` + twip(before) + `" w:after="` + twip(after) + `"`)
	if s.LineHeight > 0 {
		p.WriteString(` w:line="` + twip(s.LineHeight) + `" w:lineRule="atLeast"`)
	}
	p.WriteString(`/>`)
	left, right := s.Margin[3]+s.Padding[3], s.Margin[1]+s.Padding[1]
	if left != 0 || right != 0 || s.TextIndent != 0 {
		p.WriteString(`<w:ind w:left="` + twip(left) + `" w:right="` + twip(right) + `"`)
		if s.TextIndent < 0 {
			p.WriteString(` w:hanging="` + twip(-s.TextIndent) + `"`)
		} else if s.TextIndent > 0 {
			p.WriteString(` w:firstLine="` + twip(s.TextIndent) + `"`)
		}
		p.WriteString(`/>`)
	}
	align := s.Align
	if align == "start" {
		if s.Direction == "rtl" {
			align = "right"
		} else {
			align = "left"
		}
	}
	if align == "end" {
		if s.Direction == "rtl" {
			align = "left"
		} else {
			align = "right"
		}
	}
	if align == "justify" {
		align = "both"
	}
	if align == "left" || align == "right" || align == "center" || align == "both" {
		p.WriteString(val("jc", align))
	}
	// The paragraph mark participates in Word's line metrics even when it is
	// invisible. Leaving it at Normal 11pt inflates every small sidebar label.
	markSize := s.Size * 1.5
	if markSize <= 0 {
		markSize = w.options.FontSize * 2
	}
	markFont := fallback(s.Font, w.options.Font)
	p.WriteString(`<w:rPr><w:rFonts w:ascii="` + esc(markFont) + `" w:hAnsi="` + esc(markFont) + `" w:eastAsia="` + esc(fallback(s.EastAsia, markFont)) + `"/><w:sz w:val="` + integer(math.Max(2, markSize)) + `"/><w:szCs w:val="` + integer(math.Max(2, markSize)) + `"/></w:rPr>`)
	p.WriteString(`</w:pPr>`)
	if id, ok := w.bookmarks[b.ID]; ok {
		p.WriteString(`<w:bookmarkStart w:id="` + strconv.Itoa(id) + `" w:name="bookmark` + strconv.Itoa(id) + `"/>`)
	}
	for _, r := range b.Runs {
		run := w.run(r, math.Max(24, width-left-right))
		if r.Link != "" {
			if strings.HasPrefix(r.Link, "#") {
				if id, ok := w.bookmarks[strings.TrimPrefix(r.Link, "#")]; ok {
					run = `<w:hyperlink w:anchor="bookmark` + strconv.Itoa(id) + `">` + run + `</w:hyperlink>`
				}
			} else if u, e := url.Parse(r.Link); e == nil && (u.Scheme == "https" || u.Scheme == "http" || u.Scheme == "mailto" || u.Scheme == "tel") {
				id := w.addRel("hyperlink", r.Link, true)
				run = `<w:hyperlink r:id="` + id + `">` + run + `</w:hyperlink>`
			}
		}
		p.WriteString(run)
	}
	if id, ok := w.bookmarks[b.ID]; ok {
		p.WriteString(`<w:bookmarkEnd w:id="` + strconv.Itoa(id) + `"/>`)
	}
	if s.BreakAfter {
		p.WriteString(`<w:r><w:br w:type="page"/></w:r>`)
	}
	p.WriteString(`</w:p>`)
	return p.String()
}
func border(tag string, b Border) string {
	if b.Width <= 0 || b.Style == "none" || b.Style == "hidden" {
		return `<w:` + tag + ` w:val="nil"/>`
	}
	v := "single"
	switch b.Style {
	case "double":
		v = "double"
	case "dotted":
		v = "dotted"
	case "dashed":
		v = "dashed"
	}
	return `<w:` + tag + ` w:val="` + v + `" w:sz="` + integer(math.Min(96, math.Max(2, b.Width*6))) + `" w:space="0" w:color="` + esc(fallback(b.Color, "000000")) + `"/>`
}
func paragraphBorders(bs [4]Border) string {
	present := false
	for _, b := range bs {
		present = present || b.Width > 0 && b.Style != "none"
	}
	if !present {
		return ""
	}
	return `<w:pBdr>` + border("top", bs[0]) + border("left", bs[3]) + border("bottom", bs[2]) + border("right", bs[1]) + `</w:pBdr>`
}
func (w *writer) run(r Run, width float64) string {
	if r.Chart != nil {
		return w.chart(*r.Chart, width)
	}
	if r.Checkbox != nil {
		return w.checkbox(*r.Checkbox, r.Style, width)
	}
	if r.Image != nil {
		return w.picture(*r.Image, width)
	}
	s := r.Style
	var out strings.Builder
	out.WriteString(`<w:r><w:rPr>`)
	font := fallback(s.Font, w.options.Font)
	out.WriteString(`<w:rFonts w:ascii="` + esc(font) + `" w:hAnsi="` + esc(font) + `" w:eastAsia="` + esc(fallback(s.EastAsia, font)) + `" w:cs="` + esc(font) + `"/>`)
	out.WriteString(val("b", strconv.FormatBool(s.Bold)))
	out.WriteString(val("bCs", strconv.FormatBool(s.Bold)))
	out.WriteString(val("i", strconv.FormatBool(s.Italic)))
	out.WriteString(val("iCs", strconv.FormatBool(s.Italic)))
	if s.Strike {
		out.WriteString(`<w:strike/>`)
	}
	if s.Color != "" {
		out.WriteString(val("color", s.Color))
	}
	if s.LetterSpacing != 0 {
		out.WriteString(val("spacing", twip(s.LetterSpacing)))
	}
	size := s.Size * 1.5
	if size <= 0 {
		size = w.options.FontSize * 2
	}
	out.WriteString(val("sz", integer(math.Max(2, size))))
	out.WriteString(val("szCs", integer(math.Max(2, size))))
	if s.Underline {
		out.WriteString(val("u", "single"))
	}
	if s.Background != "" {
		out.WriteString(`<w:shd w:val="clear" w:fill="` + esc(s.Background) + `"/>`)
	}
	if s.InlineBox && s.Borders[0].Width > 0 && s.Borders[1].Width > 0 && s.Borders[2].Width > 0 && s.Borders[3].Width > 0 {
		for _, b := range s.Borders {
			if b.Width > 0 && b.Style != "none" {
				out.WriteString(strings.Replace(border("bdr", b), `w:space="0"`, `w:space="`+integer(math.Min(s.Padding[0], s.Padding[2])*.75)+`"`, 1))
				break
			}
		}
	}
	if s.NoWrap && s.Width > 0 && s.Width < 64 && strings.Trim(r.Text, "0123456789") == "" && r.Text != "" {
		out.WriteString(`<w:fitText w:val="` + twip(math.Max(1, s.Width-1)) + `"/>`)
	}
	if s.Vertical == "super" {
		out.WriteString(val("vertAlign", "superscript"))
	}
	if s.Vertical == "sub" {
		out.WriteString(val("vertAlign", "subscript"))
	}
	if s.Direction == "rtl" {
		out.WriteString(`<w:rtl/>`)
	}
	out.WriteString(`</w:rPr>`)
	if r.Break {
		out.WriteString(`<w:br/>`)
	} else {
		for i, line := range strings.Split(strings.ReplaceAll(r.Text, "\r\n", "\n"), "\n") {
			if i > 0 {
				out.WriteString(`<w:br/>`)
			}
			for j, part := range strings.Split(line, "\t") {
				if j > 0 {
					out.WriteString(`<w:tab/>`)
				}
				if part != "" {
					out.WriteString(`<w:t xml:space="preserve">` + esc(part) + `</w:t>`)
				}
			}
		}
	}
	out.WriteString(`</w:r>`)
	return out.String()
}

func (w *writer) table(b Block, width float64, depth int) string {
	var out strings.Builder
	gridCount := 1
	for _, r := range b.Rows {
		n := 0
		for _, c := range r.Cells {
			n += max(1, c.Span)
		}
		gridCount = max(gridCount, n)
	}
	// Include the extra occupied grid positions introduced by rowspans.
	occupied := map[int]int{}
	for _, row := range b.Rows {
		pos := 0
		for _, cell := range row.Cells {
			for occupied[pos] > 0 {
				pos++
			}
			span := max(1, cell.Span)
			for k := 0; k < span; k++ {
				occupied[pos+k] = max(occupied[pos+k], cell.RowSpan)
			}
			pos += span
		}
		for pos2, v := range occupied {
			gridCount = max(gridCount, pos2+1)
			occupied[pos2] = max(0, v-1)
		}
	}
	gridCount = min(gridCount, 128)
	width = math.Max(1, width-nonnegative(b.Style.Margin[1])-nonnegative(b.Style.Margin[3]))
	if b.Style.Width > 0 && b.Style.Width < width {
		width = b.Style.Width
	}
	widths := make([]float64, gridCount)
	if len(b.Grid) > 0 && len(b.Grid) <= 128 {
		gridCount = len(b.Grid)
		widths = append([]float64(nil), b.Grid...)
	}
	for _, row := range b.Rows {
		col := 0
		for _, cell := range row.Cells {
			span := max(1, cell.Span)
			if len(b.Grid) == 0 && span == 1 && col < gridCount {
				widths[col] = math.Max(widths[col], cell.Style.Width)
			}
			col += span
		}
	}
	sum := 0.0
	for _, v := range widths {
		sum += v
	}
	if sum == 0 {
		for i := range widths {
			widths[i] = width / float64(gridCount)
		}
	} else {
		missing := 0
		for _, v := range widths {
			if v == 0 {
				missing++
			}
		}
		if missing > 0 {
			avg := sum / float64(gridCount-missing)
			for i, v := range widths {
				if v == 0 {
					widths[i] = avg
					sum += avg
				}
			}
		}
		for i := range widths {
			widths[i] = widths[i] / sum * width
		}
	}
	out.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="` + twip(width) + `" w:type="dxa"/><w:jc w:val="` + tableAlign(b.Style.Align) + `"/><w:tblInd w:w="` + twip(nonnegative(b.Style.Margin[3])) + `" w:type="dxa"/><w:tblBorders><w:top w:val="nil"/><w:left w:val="nil"/><w:bottom w:val="nil"/><w:right w:val="nil"/><w:insideH w:val="nil"/><w:insideV w:val="nil"/></w:tblBorders><w:tblLayout w:type="fixed"/><w:tblCellMar><w:top w:w="0" w:type="dxa"/><w:left w:w="0" w:type="dxa"/><w:bottom w:w="0" w:type="dxa"/><w:right w:w="0" w:type="dxa"/></w:tblCellMar></w:tblPr><w:tblGrid>`)
	for _, v := range widths {
		out.WriteString(`<w:gridCol w:w="` + twip(v) + `"/>`)
	}
	out.WriteString(`</w:tblGrid>`)
	type merge struct {
		span, remaining int
		style           Style
	}
	merges := map[int]merge{}
	for _, row := range b.Rows {
		out.WriteString(`<w:tr><w:trPr>`)
		if row.Header {
			out.WriteString(`<w:tblHeader/>`)
		}
		if row.Keep || row.Header {
			out.WriteString(`<w:cantSplit/>`)
		}
		if row.Height > 0 {
			rule := "atLeast"
			if row.Exact {
				rule = "exact"
			}
			out.WriteString(`<w:trHeight w:val="` + twip(row.Height) + `" w:hRule="` + rule + `"/>`)
		}
		out.WriteString(`</w:trPr>`)
		pos := 0
		appendContinuations := func() {
			for pos < gridCount {
				m, ok := merges[pos]
				if !ok || m.remaining <= 0 {
					break
				}
				c := Cell{Span: m.span, Style: m.style, VMerge: "continue"}
				out.WriteString(w.cell(c, widths, pos, depth, b.Style.KeepNext || b.Style.Keep))
				m.remaining--
				if m.remaining == 0 {
					delete(merges, pos)
				} else {
					merges[pos] = m
				}
				pos += m.span
			}
		}
		for _, cell := range row.Cells {
			appendContinuations()
			if pos >= gridCount {
				break
			}
			cell.Span = min(max(1, cell.Span), gridCount-pos)
			if cell.RowSpan > 1 {
				cell.VMerge = "restart"
				merges[pos] = merge{cell.Span, min(10000, cell.RowSpan-1), cell.Style}
			}
			out.WriteString(w.cell(cell, widths, pos, depth, b.Style.KeepNext || b.Style.Keep || row.Header))
			pos += cell.Span
		}
		for pos < gridCount {
			if m, ok := merges[pos]; ok && m.remaining > 0 {
				appendContinuations()
			} else {
				out.WriteString(w.cell(Cell{Span: 1}, widths, pos, depth, b.Style.KeepNext || b.Style.Keep))
				pos++
			}
		}
		out.WriteString(`</w:tr>`)
	}
	out.WriteString(`</w:tbl>`)
	return out.String()
}
func tableAlign(s string) string {
	if s == "center" || s == "right" {
		return s
	}
	return "left"
}
func (w *writer) cell(c Cell, widths []float64, pos, depth int, keepNext bool) string {
	width := 0.0
	for i := pos; i < min(len(widths), pos+max(1, c.Span)); i++ {
		width += widths[i]
	}
	s := c.Style
	var out strings.Builder
	out.WriteString(`<w:tc><w:tcPr><w:tcW w:w="` + twip(width) + `" w:type="dxa"/>`)
	if c.Span > 1 {
		out.WriteString(val("gridSpan", strconv.Itoa(c.Span)))
	}
	if c.VMerge != "" {
		out.WriteString(val("vMerge", c.VMerge))
	}
	out.WriteString(`<w:tcBorders>` + border("top", s.Borders[0]) + border("left", s.Borders[3]) + border("bottom", s.Borders[2]) + border("right", s.Borders[1]) + `</w:tcBorders>`)
	if s.Background != "" {
		out.WriteString(`<w:shd w:val="clear" w:fill="` + esc(s.Background) + `"/>`)
	}
	if s.NoWrap {
		out.WriteString(`<w:noWrap/>`)
	}
	out.WriteString(`<w:tcMar><w:top w:w="` + twip(s.Padding[0]) + `" w:type="dxa"/><w:left w:w="` + twip(s.Padding[3]) + `" w:type="dxa"/><w:bottom w:w="` + twip(s.Padding[2]) + `" w:type="dxa"/><w:right w:w="` + twip(s.Padding[1]) + `" w:type="dxa"/></w:tcMar>`)
	align := "top"
	if c.VAlign == "middle" {
		align = "center"
	}
	if c.VAlign == "bottom" {
		align = "bottom"
	}
	out.WriteString(val("vAlign", align))
	out.WriteString(`</w:tcPr>`)
	if len(c.Blocks) == 0 || c.VMerge == "continue" {
		out.WriteString(spacingParagraph(0, keepNext))
	} else {
		out.WriteString(w.blocksKeeping(c.Blocks, math.Max(24, width-s.Padding[1]-s.Padding[3]), depth+1, keepNext))
	}
	out.WriteString(`</w:tc>`)
	return out.String()
}

func (w *writer) loadPicture(source string) (media, error) {
	var data []byte
	var e error
	if strings.HasPrefix(source, "data:") {
		meta, payload, ok := strings.Cut(source, ",")
		if !ok {
			return media{}, fmt.Errorf("无效 data URI")
		}
		if strings.Contains(meta, ";base64") {
			if len(payload) > 12<<20 {
				return media{}, fmt.Errorf("图片超过 8 MiB")
			}
			data, e = base64.StdEncoding.DecodeString(payload)
		} else {
			var decoded string
			decoded, e = url.PathUnescape(payload)
			data = []byte(decoded)
		}
	} else if w.execution.LoadResource != nil {
		data, _, e = w.execution.LoadResource(w.ctx, source)
	} else {
		if !w.options.AllowRemote {
			return media{}, fmt.Errorf("远程图片未启用")
		}
		transport := &http.Transport{Proxy: nil, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, e := net.SplitHostPort(address)
			if e != nil {
				return nil, e
			}
			if e = PublicResource(ctx, "https://"+net.JoinHostPort(host, port)); e != nil {
				return nil, e
			}
			ips, e := net.DefaultResolver.LookupIPAddr(ctx, host)
			if e != nil || len(ips) == 0 {
				return nil, fmt.Errorf("无法解析图片域名")
			}
			for _, a := range ips {
				if a.IP.IsLoopback() || a.IP.IsPrivate() || !a.IP.IsGlobalUnicast() {
					return nil, fmt.Errorf("拒绝私网图片地址")
				}
			}
			return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
		}}
		defer transport.CloseIdleConnections()
		client := &http.Client{Transport: transport, Timeout: 10 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 4 {
				return fmt.Errorf("图片重定向过多")
			}
			return PublicResource(req.Context(), req.URL.String())
		}}
		if e = PublicResource(w.ctx, source); e != nil {
			return media{}, e
		}
		req, err := http.NewRequestWithContext(w.ctx, "GET", source, nil)
		if err != nil {
			return media{}, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return media{}, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return media{}, fmt.Errorf("图片 HTTP %d", resp.StatusCode)
		}
		data, e = io.ReadAll(io.LimitReader(resp.Body, (8<<20)+1))
	}
	if e != nil {
		return media{}, e
	}
	if len(data) > 8<<20 {
		return media{}, fmt.Errorf("图片超过 8 MiB")
	}
	config, format, e := image.DecodeConfig(bytes.NewReader(data))
	if e != nil {
		return media{}, fmt.Errorf("不支持的图片格式: %w", e)
	}
	if config.Width*config.Height > 24_000_000 {
		return media{}, fmt.Errorf("图片像素过大")
	}
	extension := "png"
	if format == "jpeg" {
		extension = "jpg"
	} else if format != "png" {
		im, _, e := image.Decode(bytes.NewReader(data))
		if e != nil {
			return media{}, e
		}
		var b bytes.Buffer
		if e = png.Encode(&b, im); e != nil {
			return media{}, e
		}
		data = b.Bytes()
	}
	return media{Data: data, Name: fmt.Sprintf("image%d.%s", len(w.media)+1, extension), Width: config.Width, Height: config.Height}, nil
}
func (w *writer) picture(pic Picture, width float64) string {
	key := sha256.Sum256([]byte(pic.Source))
	hash := hex.EncodeToString(key[:])
	index, ok := w.imageCache[hash]
	if !ok {
		if len(w.media) >= 100 {
			w.warning("图片超过 100 张，后续图片以替代文字输出。")
			return w.run(Run{Text: "[" + pic.Alt + "]"}, width)
		}
		m, e := w.loadPicture(pic.Source)
		if e != nil {
			w.warning("图片转换失败：" + pic.Alt + "（" + e.Error() + "）")
			return w.run(Run{Text: "[" + pic.Alt + "]"}, width)
		}
		w.media = append(w.media, m)
		index = len(w.media) - 1
		w.imageCache[hash] = index
	}
	m := w.media[index]
	x, y := pic.Width, pic.Height
	if x <= 0 {
		x = float64(m.Width)
	}
	if y <= 0 {
		y = x * float64(m.Height) / float64(m.Width)
	}
	if x > width {
		y *= width / x
		x = width
	}
	_, h := w.options.PageMM()
	maxHeight := (h - 2*w.options.MarginMM) * 96 / 25.4
	if y > maxHeight {
		x *= maxHeight / y
		y = maxHeight
	}
	id := w.addRel("image", "media/"+m.Name, false)
	w.drawingID++
	cx, cy := integer(x*9525), integer(y*9525)
	geometry := `<a:prstGeom prst="rect"><a:avLst/></a:prstGeom>`
	if pic.Radius > 0 {
		geometry = `<a:prstGeom prst="roundRect"><a:avLst><a:gd name="adj" fmla="val ` + integer(math.Min(50000, pic.Radius/math.Min(x, y)*100000)) + `"/></a:avLst></a:prstGeom>`
	}
	outline := ""
	if pic.Outline.Width > 0 && pic.Outline.Style != "none" {
		outline = `<a:ln w="` + integer(pic.Outline.Width*9525) + `"><a:solidFill><a:srgbClr val="` + esc(fallback(pic.Outline.Color, "000000")) + `"/></a:solidFill></a:ln>`
	}
	return `<w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0"><wp:extent cx="` + cx + `" cy="` + cy + `"/><wp:docPr id="` + strconv.Itoa(w.drawingID) + `" name="` + esc(m.Name) + `" descr="` + esc(pic.Alt) + `"/><wp:cNvGraphicFramePr><a:graphicFrameLocks noChangeAspect="1"/></wp:cNvGraphicFramePr><a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic><pic:nvPicPr><pic:cNvPr id="0" name="` + esc(m.Name) + `"/><pic:cNvPicPr/></pic:nvPicPr><pic:blipFill><a:blip r:embed="` + id + `"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill><pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="` + cx + `" cy="` + cy + `"/></a:xfrm>` + geometry + outline + `</pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r>`
}

func (w *writer) section(columns int, gap float64, header, footer string, continuous bool) string {
	header = fallback(header, w.headerID)
	footer = fallback(footer, w.footerID)
	pw, ph := w.options.PageMM()
	var b strings.Builder
	b.WriteString(`<w:sectPr>`)
	if header != "" {
		b.WriteString(`<w:headerReference w:type="default" r:id="` + header + `"/>`)
	}
	if footer != "" {
		b.WriteString(`<w:footerReference w:type="default" r:id="` + footer + `"/>`)
	}
	if continuous {
		b.WriteString(`<w:type w:val="continuous"/>`)
	}
	b.WriteString(`<w:pgSz w:w="` + integer(pw*1440/25.4) + `" w:h="` + integer(ph*1440/25.4) + `"`)
	if w.options.Landscape {
		b.WriteString(` w:orient="landscape"`)
	}
	b.WriteString(`/><w:pgMar w:top="` + integer(w.options.MarginMM*1440/25.4) + `" w:right="` + integer(w.options.MarginMM*1440/25.4) + `" w:bottom="` + integer(w.options.MarginMM*1440/25.4) + `" w:left="` + integer(w.options.MarginMM*1440/25.4) + `" w:header="300" w:footer="300" w:gutter="0"/><w:cols w:num="` + strconv.Itoa(max(1, columns)) + `" w:space="` + twip(gap) + `"/></w:sectPr>`)
	return b.String()
}
func (w *writer) headerFooter(tag, text string, pages bool) string {
	body := `<w:p><w:pPr><w:jc w:val="center"/></w:pPr>` + w.run(Run{Text: text, Style: Style{Size: 12, Color: "64748B"}}, w.options.ContentPX())
	if pages {
		body += `<w:r><w:t xml:space="preserve">  </w:t></w:r><w:fldSimple w:instr="PAGE"><w:r><w:t>1</w:t></w:r></w:fldSimple><w:r><w:t> / </w:t></w:r><w:fldSimple w:instr="NUMPAGES"><w:r><w:t>1</w:t></w:r></w:fldSimple>`
	}
	body += `</w:p>`
	if tag == "ftr" && len(w.doc.Footer) > 0 {
		rels := w.rels
		w.rels = nil
		body = w.blocks(w.doc.Footer, w.options.ContentPX(), 1) + body
		if len(w.rels) > 0 {
			w.extraParts["word/_rels/footer1.xml.rels"] = []byte(w.relationships())
		}
		w.rels = rels
	}
	return xmlHead + `<w:` + tag + ` ` + wordNS + `>` + body + `</w:` + tag + `>`
}
func (w *writer) styles() string {
	size := integer(w.options.FontSize * 2)
	font := esc(w.options.Font)
	var b strings.Builder
	b.WriteString(xmlHead + `<w:styles ` + wordNS + `><w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="` + font + `" w:hAnsi="` + font + `" w:eastAsia="` + font + `" w:cs="` + font + `"/><w:sz w:val="` + size + `"/><w:szCs w:val="` + size + `"/><w:lang w:val="en-US" w:eastAsia="zh-CN"/></w:rPr></w:rPrDefault><w:pPrDefault><w:pPr><w:widowControl/><w:spacing w:before="0" w:after="0"/></w:pPr></w:pPrDefault></w:docDefaults><w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/><w:qFormat/></w:style>`)
	for i := 1; i <= 6; i++ {
		b.WriteString(`<w:style w:type="paragraph" w:styleId="Heading` + strconv.Itoa(i) + `"><w:name w:val="heading ` + strconv.Itoa(i) + `"/><w:basedOn w:val="Normal"/><w:next w:val="Normal"/><w:uiPriority w:val="9"/><w:qFormat/><w:pPr><w:keepNext/><w:keepLines/><w:outlineLvl w:val="` + strconv.Itoa(i-1) + `"/></w:pPr></w:style>`)
	}
	b.WriteString(`</w:styles>`)
	return b.String()
}
func (w *writer) numbering() string {
	var b strings.Builder
	b.WriteString(xmlHead + `<w:numbering ` + wordNS + `>`)
	ids := make([]int, 0, len(w.lists))
	for id := range w.lists {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		s := w.lists[id]
		b.WriteString(`<w:abstractNum w:abstractNumId="` + strconv.Itoa(id) + `"><w:multiLevelType w:val="multilevel"/>`)
		for level := 0; level < 9; level++ {
			format, text := "bullet", "•"
			if s.Ordered {
				format = "decimal"
				switch s.Format {
				case "lower-alpha", "lower-latin":
					format = "lowerLetter"
				case "upper-alpha", "upper-latin":
					format = "upperLetter"
				case "lower-roman":
					format = "lowerRoman"
				case "upper-roman":
					format = "upperRoman"
				}
				text = fmt.Sprintf("%%%d.", level+1)
			} else if s.Format == "circle" {
				text = "○"
			} else if s.Format == "square" {
				text = "▪"
			}
			b.WriteString(`<w:lvl w:ilvl="` + strconv.Itoa(level) + `"><w:start w:val="` + strconv.Itoa(s.Start) + `"/><w:numFmt w:val="` + format + `"/><w:lvlText w:val="` + esc(text) + `"/><w:lvlJc w:val="left"/><w:pPr><w:tabs><w:tab w:val="num" w:pos="` + strconv.Itoa(360*(level+1)) + `"/></w:tabs><w:ind w:left="` + strconv.Itoa(360*(level+1)) + `" w:hanging="240"/></w:pPr></w:lvl>`)
		}
		b.WriteString(`</w:abstractNum>`)
	}
	for _, id := range ids {
		b.WriteString(`<w:num w:numId="` + strconv.Itoa(id) + `"><w:abstractNumId w:val="` + strconv.Itoa(id) + `"/></w:num>`)
	}
	b.WriteString(`</w:numbering>`)
	return b.String()
}
func (w *writer) relationships() string {
	var b strings.Builder
	b.WriteString(xmlHead + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for _, r := range w.rels {
		b.WriteString(`<Relationship Id="` + r.ID + `" Type="` + relBase + r.Kind + `" Target="` + esc(r.Target) + `"`)
		if r.External {
			b.WriteString(` TargetMode="External"`)
		}
		b.WriteString(`/>`)
	}
	b.WriteString(`</Relationships>`)
	return b.String()
}
func (w *writer) contentTypes(header, footer bool) string {
	var b strings.Builder
	b.WriteString(xmlHead + `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Default Extension="png" ContentType="image/png"/><Default Extension="jpg" ContentType="image/jpeg"/>`)
	for _, part := range []struct{ name, kind string }{{"document", "document.main"}, {"styles", "styles"}, {"settings", "settings"}, {"fontTable", "fontTable"}, {"numbering", "numbering"}} {
		b.WriteString(`<Override PartName="/word/` + part.name + `.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.` + part.kind + `+xml"/>`)
	}
	if header {
		b.WriteString(`<Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>`)
	}
	if footer {
		b.WriteString(`<Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>`)
	}
	if w.charts > 0 {
		b.WriteString(`<Default Extension="xlsx" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"/>`)
	}
	for i := 1; i <= w.charts; i++ {
		b.WriteString(`<Override PartName="/word/charts/chart` + strconv.Itoa(i) + `.xml" ContentType="application/vnd.openxmlformats-officedocument.drawingml.chart+xml"/>`)
	}
	b.WriteString(`<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/><Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/></Types>`)
	return b.String()
}

func SafeFilename(name string) string {
	var b strings.Builder
	for _, r := range strings.TrimSuffix(name, ".docx") {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == ' ' || r == '.' {
			b.WriteRune(r)
		}
	}
	s := strings.Trim(b.String(), " .")
	if s == "" {
		s = "document"
	}
	if len([]rune(s)) > 80 {
		s = string([]rune(s)[:80])
	}
	return s + ".docx"
}
